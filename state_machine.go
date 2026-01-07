package stateless

import (
	"context"
	"fmt"
	"sync"
)

// FiringMode determines how the state machine handles multiple trigger fires.
type FiringMode int

const (
	// FiringImmediate causes triggers to be processed immediately (synchronously).
	// This is the default mode.
	FiringImmediate FiringMode = iota

	// FiringQueued causes triggers to be queued and processed one at a time.
	// This ensures only one trigger is being processed at any time.
	FiringQueued
)

// StateMachine represents a state machine that can transition between states based on triggers.
type StateMachine[TState, TTrigger comparable] struct {
	// stateAccessor is used to retrieve the current state.
	stateAccessor func() TState

	// stateMutator is used to set the current state.
	stateMutator func(TState)

	// stateRepresentations contains the configuration for each state.
	stateRepresentations map[TState]*StateRepresentation[TState, TTrigger]

	// triggerConfigurations contains the configuration for parameterized triggers.
	triggerConfigurations map[TTrigger]*TriggerWithParameters[TTrigger]

	// unhandledTriggerAction is called when a trigger is fired but not handled.
	unhandledTriggerAction func(state TState, trigger TTrigger, unmetGuards []string)

	// onTransitionedEvent is called when a transition is completed.
	onTransitionedEvent *OnTransitionedEvent[TState, TTrigger]

	// onTransitionCompletedEvent is called after all transition actions are executed.
	onTransitionCompletedEvent *OnTransitionedEvent[TState, TTrigger]

	// firingMode determines how triggers are processed.
	firingMode FiringMode

	// eventQueue holds queued events when using FiringQueued mode.
	eventQueue []queuedEvent[TState, TTrigger]

	// firing indicates if the state machine is currently processing a trigger.
	firing bool

	// mutex protects the state machine from concurrent access.
	mutex sync.Mutex

	// retainSynchronizationContext indicates if synchronization context should be retained.
	retainSynchronizationContext bool

	// isActive indicates if the state machine has been activated.
	isActive bool

	// initialState stores the initial state of the state machine.
	initialState TState
}

// queuedEvent represents an event waiting to be processed.
type queuedEvent[TState, TTrigger comparable] struct {
	trigger TTrigger
	args    any
	ctx     context.Context
}

// OnTransitionedEvent handles transition event callbacks.
type OnTransitionedEvent[TState, TTrigger comparable] struct {
	handlers []func(internalTransition[TState, TTrigger])
	mutex    sync.RWMutex
}

// NewOnTransitionedEvent creates a new OnTransitionedEvent.
func NewOnTransitionedEvent[TState, TTrigger comparable]() *OnTransitionedEvent[TState, TTrigger] {
	return &OnTransitionedEvent[TState, TTrigger]{}
}

// Register adds a handler to the event.
func (e *OnTransitionedEvent[TState, TTrigger]) Register(handler func(internalTransition[TState, TTrigger])) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.handlers = append(e.handlers, handler)
}

// UnregisterAll removes all handlers from the event.
func (e *OnTransitionedEvent[TState, TTrigger]) UnregisterAll() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.handlers = nil
}

// Invoke calls all registered handlers.
func (e *OnTransitionedEvent[TState, TTrigger]) Invoke(transition internalTransition[TState, TTrigger]) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	for _, handler := range e.handlers {
		handler(transition)
	}
}

// NewStateMachine creates a new state machine with the specified initial state.
func NewStateMachine[TState, TTrigger comparable](initialState TState) *StateMachine[TState, TTrigger] {
	var state TState = initialState
	return NewStateMachineWithExternalStorage[TState, TTrigger](
		func() TState { return state },
		func(s TState) { state = s },
	)
}

// NewStateMachineWithMode creates a new state machine with the specified initial state and firing mode.
func NewStateMachineWithMode[TState, TTrigger comparable](initialState TState, firingMode FiringMode) *StateMachine[TState, TTrigger] {
	sm := NewStateMachine[TState, TTrigger](initialState)
	sm.firingMode = firingMode
	return sm
}

// NewStateMachineWithExternalStorage creates a new state machine with external state storage.
func NewStateMachineWithExternalStorage[TState, TTrigger comparable](
	stateAccessor func() TState,
	stateMutator func(TState),
) *StateMachine[TState, TTrigger] {
	return &StateMachine[TState, TTrigger]{
		stateAccessor:              stateAccessor,
		stateMutator:               stateMutator,
		stateRepresentations:       make(map[TState]*StateRepresentation[TState, TTrigger]),
		triggerConfigurations:      make(map[TTrigger]*TriggerWithParameters[TTrigger]),
		onTransitionedEvent:        NewOnTransitionedEvent[TState, TTrigger](),
		onTransitionCompletedEvent: NewOnTransitionedEvent[TState, TTrigger](),
		firingMode:                 FiringImmediate,
		initialState:               stateAccessor(),
	}
}

// NewStateMachineWithExternalStorageAndMode creates a new state machine with external state storage
// and the specified firing mode.
func NewStateMachineWithExternalStorageAndMode[TState, TTrigger comparable](
	stateAccessor func() TState,
	stateMutator func(TState),
	firingMode FiringMode,
) *StateMachine[TState, TTrigger] {
	sm := NewStateMachineWithExternalStorage[TState, TTrigger](stateAccessor, stateMutator)
	sm.firingMode = firingMode
	return sm
}

// State returns the current state.
func (sm *StateMachine[TState, TTrigger]) State() TState {
	return sm.stateAccessor()
}

// Configure begins configuration of a state.
func (sm *StateMachine[TState, TTrigger]) Configure(state TState) *StateConfiguration[TState, TTrigger] {
	return NewStateConfiguration(
		sm.getRepresentation(state),
		sm.getRepresentation,
		sm.triggerConfigurations,
	)
}

// SetTriggerParameters configures a trigger to require specific parameter types.
func (sm *StateMachine[TState, TTrigger]) SetTriggerParameters(trigger TTrigger, argTypes ...any) *TriggerWithParameters[TTrigger] {
	config := NewTriggerWithParameters(trigger)
	sm.triggerConfigurations[trigger] = config
	return config
}

// Fire fires a trigger with optional args (should be a struct or nil).
func (sm *StateMachine[TState, TTrigger]) Fire(trigger TTrigger, args any) error {
	return sm.FireCtx(context.Background(), trigger, args)
}

// FireCtx fires a trigger with a context and optional args.
func (sm *StateMachine[TState, TTrigger]) FireCtx(ctx context.Context, trigger TTrigger, args any) error {
	sm.mutex.Lock()

	if sm.firingMode == FiringQueued {
		sm.eventQueue = append(sm.eventQueue, queuedEvent[TState, TTrigger]{
			trigger: trigger,
			args:    args,
			ctx:     ctx,
		})

		if sm.firing {
			sm.mutex.Unlock()
			return nil
		}

		sm.firing = true
		sm.mutex.Unlock()

		for {
			sm.mutex.Lock()
			if len(sm.eventQueue) == 0 {
				sm.firing = false
				sm.mutex.Unlock()
				return nil
			}
			event := sm.eventQueue[0]
			sm.eventQueue = sm.eventQueue[1:]
			sm.mutex.Unlock()

			if err := sm.internalFire(event.ctx, event.trigger, event.args); err != nil {
				sm.mutex.Lock()
				sm.firing = false
				sm.mutex.Unlock()
				return err
			}
		}
	}

	sm.mutex.Unlock()
	return sm.internalFire(ctx, trigger, args)
}

// internalFire processes a single trigger.
func (sm *StateMachine[TState, TTrigger]) internalFire(ctx context.Context, trigger TTrigger, args any) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	source := sm.State()
	representation := sm.getRepresentation(source)

	// Try to find a handler for the trigger
	result := representation.TryFindHandler(trigger, args)
	if result == nil || result.Handler == nil {
		return sm.handleUnhandledTrigger(source, trigger, result)
	}

	handler := result.Handler

	// Handle different types of trigger behaviours
	switch behaviour := handler.(type) {
	case *TransitioningTriggerBehaviour[TState, TTrigger]:
		destination := behaviour.Destination
		transition := internalTransition[TState, TTrigger]{
			Source:      source,
			Destination: destination,
			Trigger:     trigger,
			Args:        args,
		}

		// Execute exit actions
		if err := representation.Exit(transition); err != nil {
			return err
		}

		// Update state
		sm.stateMutator(destination)

		// Fire transition event
		sm.onTransitionedEvent.Invoke(transition)

		// Get destination representation
		destRepresentation := sm.getRepresentation(destination)

		// Execute entry actions
		if err := destRepresentation.Enter(transition); err != nil {
			return err
		}

		// Handle initial transition if destination has one
		if destRepresentation.HasInitialTransition() {
			initialTarget := destRepresentation.InitialTransitionTarget()
			initialTransition := internalTransition[TState, TTrigger]{
				Source:      destination,
				Destination: initialTarget,
				Trigger:     trigger,
				Args:        args,
				isInitial:   true,
			}

			// Update state to initial target
			sm.stateMutator(initialTarget)

			// Execute entry actions for initial target
			initialTargetRepresentation := sm.getRepresentation(initialTarget)
			if err := initialTargetRepresentation.Enter(initialTransition); err != nil {
				return err
			}
		}

		// Fire transition completed event
		finalTransition := internalTransition[TState, TTrigger]{
			Source:      source,
			Destination: sm.State(),
			Trigger:     trigger,
			Args:        args,
		}
		sm.onTransitionCompletedEvent.Invoke(finalTransition)

	case *ReentryTriggerBehaviour[TState, TTrigger]:
		destination := behaviour.Destination
		transition := internalTransition[TState, TTrigger]{
			Source:      source,
			Destination: destination,
			Trigger:     trigger,
			Args:        args,
		}

		// Execute exit actions (reentry still fires exit/entry)
		if err := representation.Exit(transition); err != nil {
			return err
		}

		sm.stateMutator(destination)
		sm.onTransitionedEvent.Invoke(transition)

		// Execute entry actions
		destRepresentation := sm.getRepresentation(destination)
		if err := destRepresentation.Enter(transition); err != nil {
			return err
		}

		sm.onTransitionCompletedEvent.Invoke(transition)

	case *IgnoredTriggerBehaviour[TState, TTrigger]:
		// Trigger is ignored, do nothing

	case *DynamicTriggerBehaviour[TState, TTrigger]:
		destination := behaviour.GetDestinationState(args)
		transition := internalTransition[TState, TTrigger]{
			Source:      source,
			Destination: destination,
			Trigger:     trigger,
			Args:        args,
		}

		// Execute exit actions
		if err := representation.Exit(transition); err != nil {
			return err
		}

		// Update state
		sm.stateMutator(destination)

		// Fire transition event
		sm.onTransitionedEvent.Invoke(transition)

		// Execute entry actions
		destRepresentation := sm.getRepresentation(destination)
		if err := destRepresentation.Enter(transition); err != nil {
			return err
		}

		// Handle initial transition if destination has one
		if destRepresentation.HasInitialTransition() {
			initialTarget := destRepresentation.InitialTransitionTarget()
			initialTransition := internalTransition[TState, TTrigger]{
				Source:      destination,
				Destination: initialTarget,
				Trigger:     trigger,
				Args:        args,
				isInitial:   true,
			}
			sm.stateMutator(initialTarget)

			initialTargetRepresentation := sm.getRepresentation(initialTarget)
			if err := initialTargetRepresentation.Enter(initialTransition); err != nil {
				return err
			}
		}

		finalTransition := internalTransition[TState, TTrigger]{
			Source:      source,
			Destination: sm.State(),
			Trigger:     trigger,
			Args:        args,
		}
		sm.onTransitionCompletedEvent.Invoke(finalTransition)

	case InternalTriggerBehaviour[TState, TTrigger]:
		transition := internalTransition[TState, TTrigger]{
			Source:      source,
			Destination: source,
			Trigger:     trigger,
			Args:        args,
		}
		if err := behaviour.Execute(transition); err != nil {
			return err
		}
		// Internal transitions don't fire transition events

	default:
		return &InvalidOperationError{Message: fmt.Sprintf("unknown trigger behaviour type: %T", handler)}
	}

	return nil
}

// handleUnhandledTrigger handles a trigger that has no valid handler.
func (sm *StateMachine[TState, TTrigger]) handleUnhandledTrigger(state TState, trigger TTrigger, result *TriggerBehaviourResult[TState, TTrigger]) error {
	var unmetGuards []string
	if result != nil {
		unmetGuards = result.UnmetGuardConditions
	}

	if sm.unhandledTriggerAction != nil {
		sm.unhandledTriggerAction(state, trigger, unmetGuards)
		return nil
	}

	// Get permitted triggers for the error message
	representation := sm.getRepresentation(state)
	permittedTriggers := representation.GetPermittedTriggers(nil)

	// Convert to any slice for the error
	permitted := make([]any, len(permittedTriggers))
	for i, t := range permittedTriggers {
		permitted[i] = t
	}

	return &InvalidTransitionError{
		Trigger:           trigger,
		State:             state,
		UnmetGuards:       unmetGuards,
		PermittedTriggers: permitted,
	}
}

// OnUnhandledTrigger registers a callback that will be called when a trigger is fired
// but no valid transition exists.
func (sm *StateMachine[TState, TTrigger]) OnUnhandledTrigger(action func(state TState, trigger TTrigger, unmetGuards []string)) {
	sm.unhandledTriggerAction = action
}

// OnTransitioned registers a callback that will be called when a transition is completed.
// The callback receives a typed Transition with the specified TArgs type.
func OnTransitioned[TState, TTrigger comparable, TArgs any](
	sm *StateMachine[TState, TTrigger],
	action func(Transition[TState, TTrigger, TArgs]),
) {
	sm.onTransitionedEvent.Register(func(t internalTransition[TState, TTrigger]) {
		action(toTypedTransition[TState, TTrigger, TArgs](t))
	})
}

// OnTransitionCompleted registers a callback that will be called after all transition actions are executed.
// The callback receives a typed Transition with the specified TArgs type.
func OnTransitionCompleted[TState, TTrigger comparable, TArgs any](
	sm *StateMachine[TState, TTrigger],
	action func(Transition[TState, TTrigger, TArgs]),
) {
	sm.onTransitionCompletedEvent.Register(func(t internalTransition[TState, TTrigger]) {
		action(toTypedTransition[TState, TTrigger, TArgs](t))
	})
}

// UnregisterAllTransitionedCallbacks removes all OnTransitioned callbacks.
func (sm *StateMachine[TState, TTrigger]) UnregisterAllTransitionedCallbacks() {
	sm.onTransitionedEvent.UnregisterAll()
}

// UnregisterAllTransitionCompletedCallbacks removes all OnTransitionCompleted callbacks.
func (sm *StateMachine[TState, TTrigger]) UnregisterAllTransitionCompletedCallbacks() {
	sm.onTransitionCompletedEvent.UnregisterAll()
}

// UnregisterAllCallbacks removes all registered callbacks (OnTransitioned and OnTransitionCompleted).
func (sm *StateMachine[TState, TTrigger]) UnregisterAllCallbacks() {
	sm.onTransitionedEvent.UnregisterAll()
	sm.onTransitionCompletedEvent.UnregisterAll()
	sm.unhandledTriggerAction = nil
}

// Activate activates the state machine.
func (sm *StateMachine[TState, TTrigger]) Activate() error {
	if sm.isActive {
		return nil
	}

	currentRepresentation := sm.getRepresentation(sm.State())
	if err := currentRepresentation.Activate(); err != nil {
		return err
	}

	sm.isActive = true
	return nil
}

// Deactivate deactivates the state machine.
func (sm *StateMachine[TState, TTrigger]) Deactivate() error {
	if !sm.isActive {
		return nil
	}

	currentRepresentation := sm.getRepresentation(sm.State())
	if err := currentRepresentation.Deactivate(); err != nil {
		return err
	}

	sm.isActive = false
	return nil
}

// IsInState returns true if the current state is the specified state or a substate of it.
func (sm *StateMachine[TState, TTrigger]) IsInState(state TState) bool {
	currentRepresentation := sm.getRepresentation(sm.State())
	return currentRepresentation.IsIncludedIn(state)
}

// CanFire returns true if the specified trigger can be fired from the current state.
func (sm *StateMachine[TState, TTrigger]) CanFire(trigger TTrigger, args any) bool {
	return sm.getRepresentation(sm.State()).CanHandle(trigger, args)
}

// GetPermittedTriggers returns the triggers that can be fired from the current state.
func (sm *StateMachine[TState, TTrigger]) GetPermittedTriggers(args any) []TTrigger {
	return sm.getRepresentation(sm.State()).GetPermittedTriggers(args)
}

// GetDetailedPermittedTriggers returns detailed information about permitted triggers.
func (sm *StateMachine[TState, TTrigger]) GetDetailedPermittedTriggers(args any) []TriggerDetails[TState, TTrigger] {
	triggers := sm.GetPermittedTriggers(args)
	details := make([]TriggerDetails[TState, TTrigger], len(triggers))
	for i, trigger := range triggers {
		details[i] = NewTriggerDetails[TState](trigger, sm.triggerConfigurations)
	}
	return details
}

// getRepresentation gets or creates the representation for a state.
func (sm *StateMachine[TState, TTrigger]) getRepresentation(state TState) *StateRepresentation[TState, TTrigger] {
	representation, exists := sm.stateRepresentations[state]
	if !exists {
		representation = NewStateRepresentation[TState, TTrigger](state)
		sm.stateRepresentations[state] = representation
	}
	return representation
}

// GetInfo returns information about the state machine configuration for introspection.
func (sm *StateMachine[TState, TTrigger]) GetInfo() *StateMachineInfo {
	// Build state info map first
	stateInfos := make(map[TState]*StateInfo)

	// Create StateInfo for each state
	for state, rep := range sm.stateRepresentations {
		stateInfos[state] = sm.createStateInfo(rep)
	}

	// Add relationships (substates, superstates, transitions)
	for state, rep := range sm.stateRepresentations {
		sm.addStateRelationships(stateInfos[state], rep, stateInfos)
	}

	// Convert to slice
	states := make([]*StateInfo, 0, len(stateInfos))
	for _, info := range stateInfos {
		states = append(states, info)
	}

	// Find initial state info
	var initialStateInfo *StateInfo
	if info, ok := stateInfos[sm.initialState]; ok {
		initialStateInfo = info
	}

	return &StateMachineInfo{
		InitialState: initialStateInfo,
		States:       states,
		StateType:    fmt.Sprintf("%T", sm.initialState),
		TriggerType:  fmt.Sprintf("%T", *new(TTrigger)),
	}
}

// createStateInfo creates a StateInfo from a StateRepresentation.
func (sm *StateMachine[TState, TTrigger]) createStateInfo(rep *StateRepresentation[TState, TTrigger]) *StateInfo {
	// Gather ignored triggers
	var ignoredTriggers []IgnoredTransitionInfo
	for trigger, behaviours := range rep.TriggerBehaviours() {
		for _, behaviour := range behaviours {
			if _, ok := behaviour.(*IgnoredTriggerBehaviour[TState, TTrigger]); ok {
				ignoredTriggers = append(ignoredTriggers, IgnoredTransitionInfo{
					transitionInfoBase: transitionInfoBase{
						Trigger:         NewTriggerInfo(trigger),
						GuardConditions: convertGuardConditions(behaviour.GetGuard().Conditions),
					},
				})
			}
		}
	}

	// Gather entry actions
	entryActions := make([]ActionInfo, len(rep.EntryActions()))
	for i, action := range rep.EntryActions() {
		var fromTrigger any
		if ft := action.GetFromTrigger(); ft != nil {
			fromTrigger = *ft
		}
		entryActions[i] = NewActionInfo(action.GetDescription(), fromTrigger)
	}

	// Gather activate actions
	activateActions := make([]InvocationInfo, len(rep.ActivateActions()))
	for i, action := range rep.ActivateActions() {
		activateActions[i] = action.GetDescription()
	}

	// Gather deactivate actions
	deactivateActions := make([]InvocationInfo, len(rep.DeactivateActions()))
	for i, action := range rep.DeactivateActions() {
		deactivateActions[i] = action.GetDescription()
	}

	// Gather exit actions
	exitActions := make([]InvocationInfo, len(rep.ExitActions()))
	for i, action := range rep.ExitActions() {
		exitActions[i] = action.GetDescription()
	}

	return &StateInfo{
		UnderlyingState:   rep.UnderlyingState(),
		IgnoredTriggers:   ignoredTriggers,
		EntryActions:      entryActions,
		ActivateActions:   activateActions,
		DeactivateActions: deactivateActions,
		ExitActions:       exitActions,
	}
}

// addStateRelationships adds relationships to a StateInfo.
func (sm *StateMachine[TState, TTrigger]) addStateRelationships(
	info *StateInfo,
	rep *StateRepresentation[TState, TTrigger],
	stateInfos map[TState]*StateInfo,
) {
	// Add superstate
	if rep.Superstate() != nil {
		if superstateInfo, ok := stateInfos[rep.Superstate().UnderlyingState()]; ok {
			info.Superstate = superstateInfo
		}
	}

	// Add substates
	for _, substate := range rep.GetSubstates() {
		if substateInfo, ok := stateInfos[substate.UnderlyingState()]; ok {
			info.Substates = append(info.Substates, substateInfo)
		}
	}

	// Add fixed transitions
	for trigger, behaviours := range rep.TriggerBehaviours() {
		for _, behaviour := range behaviours {
			switch b := behaviour.(type) {
			case *TransitioningTriggerBehaviour[TState, TTrigger]:
				if destInfo, ok := stateInfos[b.Destination]; ok {
					info.FixedTransitions = append(info.FixedTransitions, FixedTransitionInfo{
						transitionInfoBase: transitionInfoBase{
							Trigger:              NewTriggerInfo(trigger),
							GuardConditions:      convertGuardConditions(behaviour.GetGuard().Conditions),
							IsInternalTransition: false,
						},
						DestinationState: destInfo,
					})
				}
			case *ReentryTriggerBehaviour[TState, TTrigger]:
				if destInfo, ok := stateInfos[b.Destination]; ok {
					info.FixedTransitions = append(info.FixedTransitions, FixedTransitionInfo{
						transitionInfoBase: transitionInfoBase{
							Trigger:              NewTriggerInfo(trigger),
							GuardConditions:      convertGuardConditions(behaviour.GetGuard().Conditions),
							IsInternalTransition: false,
						},
						DestinationState: destInfo,
					})
				}
			case InternalTriggerBehaviour[TState, TTrigger]:
				if destInfo, ok := stateInfos[rep.UnderlyingState()]; ok {
					info.FixedTransitions = append(info.FixedTransitions, FixedTransitionInfo{
						transitionInfoBase: transitionInfoBase{
							Trigger:              NewTriggerInfo(trigger),
							GuardConditions:      convertGuardConditions(behaviour.GetGuard().Conditions),
							IsInternalTransition: true,
						},
						DestinationState: destInfo,
					})
				}
			case *DynamicTriggerBehaviour[TState, TTrigger]:
				info.DynamicTransitions = append(info.DynamicTransitions, b.TransitionInfo)
			}
		}
	}
}

// convertGuardConditions converts GuardConditions to InvocationInfos.
func convertGuardConditions(conditions []GuardCondition) []InvocationInfo {
	result := make([]InvocationInfo, len(conditions))
	for i, c := range conditions {
		result[i] = c.MethodDescription()
	}
	return result
}

// String returns a string representation of the current state.
func (sm *StateMachine[TState, TTrigger]) String() string {
	return fmt.Sprintf("StateMachine { State = %v }", sm.State())
}
