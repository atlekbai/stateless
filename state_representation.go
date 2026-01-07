package stateless

import (
	"fmt"
)

// StateRepresentation models the behaviour of a state.
type StateRepresentation[TState, TTrigger comparable] struct {
	state TState

	// superstate is the parent state (nil if this is a root state).
	superstate *StateRepresentation[TState, TTrigger]

	// substates are the child states of this state.
	substates []*StateRepresentation[TState, TTrigger]

	// triggerBehaviours maps triggers to their behaviours.
	triggerBehaviours map[TTrigger][]TriggerBehaviour[TState, TTrigger]

	// entryActions are executed when entering this state.
	entryActions []EntryActionBehaviour[TState, TTrigger]

	// exitActions are executed when leaving this state.
	exitActions []ExitActionBehaviour[TState, TTrigger]

	// activateActions are executed when this state is activated.
	activateActions []ActivateActionBehaviour[TState]

	// deactivateActions are executed when this state is deactivated.
	deactivateActions []DeactivateActionBehaviour[TState]

	// hasInitialTransition indicates if this state has an initial transition configured.
	hasInitialTransition bool

	// initialTransitionTarget is the target state for the initial transition.
	initialTransitionTarget TState
}

// NewStateRepresentation creates a new state representation.
func NewStateRepresentation[TState, TTrigger comparable](state TState) *StateRepresentation[TState, TTrigger] {
	return &StateRepresentation[TState, TTrigger]{
		state:             state,
		triggerBehaviours: make(map[TTrigger][]TriggerBehaviour[TState, TTrigger]),
	}
}

// UnderlyingState returns the state this representation models.
func (sr *StateRepresentation[TState, TTrigger]) UnderlyingState() TState {
	return sr.state
}

// Superstate returns the parent state, if any.
func (sr *StateRepresentation[TState, TTrigger]) Superstate() *StateRepresentation[TState, TTrigger] {
	return sr.superstate
}

// SetSuperstate sets the parent state.
func (sr *StateRepresentation[TState, TTrigger]) SetSuperstate(superstate *StateRepresentation[TState, TTrigger]) {
	sr.superstate = superstate
}

// GetSubstates returns the substates of this state.
func (sr *StateRepresentation[TState, TTrigger]) GetSubstates() []*StateRepresentation[TState, TTrigger] {
	return sr.substates
}

// AddSubstate adds a substate to this state.
func (sr *StateRepresentation[TState, TTrigger]) AddSubstate(substate *StateRepresentation[TState, TTrigger]) {
	sr.substates = append(sr.substates, substate)
}

// TriggerBehaviours returns the trigger behaviours map.
func (sr *StateRepresentation[TState, TTrigger]) TriggerBehaviours() map[TTrigger][]TriggerBehaviour[TState, TTrigger] {
	return sr.triggerBehaviours
}

// EntryActions returns the entry actions.
func (sr *StateRepresentation[TState, TTrigger]) EntryActions() []EntryActionBehaviour[TState, TTrigger] {
	return sr.entryActions
}

// ExitActions returns the exit actions.
func (sr *StateRepresentation[TState, TTrigger]) ExitActions() []ExitActionBehaviour[TState, TTrigger] {
	return sr.exitActions
}

// ActivateActions returns the activate actions.
func (sr *StateRepresentation[TState, TTrigger]) ActivateActions() []ActivateActionBehaviour[TState] {
	return sr.activateActions
}

// DeactivateActions returns the deactivate actions.
func (sr *StateRepresentation[TState, TTrigger]) DeactivateActions() []DeactivateActionBehaviour[TState] {
	return sr.deactivateActions
}

// HasInitialTransition returns true if this state has an initial transition configured.
func (sr *StateRepresentation[TState, TTrigger]) HasInitialTransition() bool {
	return sr.hasInitialTransition
}

// InitialTransitionTarget returns the target state for the initial transition.
func (sr *StateRepresentation[TState, TTrigger]) InitialTransitionTarget() TState {
	return sr.initialTransitionTarget
}

// SetInitialTransition sets the initial transition for this state.
func (sr *StateRepresentation[TState, TTrigger]) SetInitialTransition(target TState) {
	sr.hasInitialTransition = true
	sr.initialTransitionTarget = target
}

// CanHandle returns true if this state can handle the specified trigger.
func (sr *StateRepresentation[TState, TTrigger]) CanHandle(trigger TTrigger, args ...any) bool {
	result := sr.TryFindHandler(trigger, args...)
	return result != nil && result.Handler != nil
}

// TryFindHandler attempts to find a handler for the specified trigger.
func (sr *StateRepresentation[TState, TTrigger]) TryFindHandler(trigger TTrigger, args ...any) *TriggerBehaviourResult[TState, TTrigger] {
	result := sr.TryFindLocalHandler(trigger, args...)
	if result == nil && sr.superstate != nil {
		result = sr.superstate.TryFindHandler(trigger, args...)
	}
	return result
}

// TryFindLocalHandler attempts to find a local handler for the specified trigger.
func (sr *StateRepresentation[TState, TTrigger]) TryFindLocalHandler(trigger TTrigger, args ...any) *TriggerBehaviourResult[TState, TTrigger] {
	behaviours, exists := sr.triggerBehaviours[trigger]
	if !exists {
		return nil
	}

	// Find all possible handlers that meet guard conditions
	var possibleBehaviours []TriggerBehaviour[TState, TTrigger]
	for _, behaviour := range behaviours {
		if behaviour.GuardConditionsMet(args...) {
			possibleBehaviours = append(possibleBehaviours, behaviour)
		}
	}

	if len(possibleBehaviours) > 1 {
		// Multiple handlers met guard conditions - this is typically an error
		// Return nil to indicate ambiguity (the caller should handle this)
		return nil
	}

	if len(possibleBehaviours) == 1 {
		return &TriggerBehaviourResult[TState, TTrigger]{
			Handler:              possibleBehaviours[0],
			UnmetGuardConditions: nil,
		}
	}

	// No handlers met guard conditions, return information about unmet guards
	var unmetGuards []string
	for _, behaviour := range behaviours {
		unmetGuards = append(unmetGuards, behaviour.UnmetGuardConditions(args...)...)
	}

	return &TriggerBehaviourResult[TState, TTrigger]{
		Handler:              nil,
		UnmetGuardConditions: unmetGuards,
	}
}

// AddTriggerBehaviour adds a trigger behaviour to this state.
func (sr *StateRepresentation[TState, TTrigger]) AddTriggerBehaviour(behaviour TriggerBehaviour[TState, TTrigger]) {
	trigger := behaviour.GetTrigger()
	sr.triggerBehaviours[trigger] = append(sr.triggerBehaviours[trigger], behaviour)
}

// AddEntryAction adds an entry action to this state.
func (sr *StateRepresentation[TState, TTrigger]) AddEntryAction(action EntryActionBehaviour[TState, TTrigger]) {
	sr.entryActions = append(sr.entryActions, action)
}

// AddExitAction adds an exit action to this state.
func (sr *StateRepresentation[TState, TTrigger]) AddExitAction(action ExitActionBehaviour[TState, TTrigger]) {
	sr.exitActions = append(sr.exitActions, action)
}

// AddActivateAction adds an activate action to this state.
func (sr *StateRepresentation[TState, TTrigger]) AddActivateAction(action ActivateActionBehaviour[TState]) {
	sr.activateActions = append(sr.activateActions, action)
}

// AddDeactivateAction adds a deactivate action to this state.
func (sr *StateRepresentation[TState, TTrigger]) AddDeactivateAction(action DeactivateActionBehaviour[TState]) {
	sr.deactivateActions = append(sr.deactivateActions, action)
}

// Enter executes entry actions for this state.
func (sr *StateRepresentation[TState, TTrigger]) Enter(transition Transition[TState, TTrigger], args ...any) error {
	if transition.IsReentry() {
		return sr.ExecuteEntryActions(transition, args...)
	}

	if !sr.Includes(transition.Source) {
		if sr.superstate != nil {
			if err := sr.superstate.Enter(transition, args...); err != nil {
				return err
			}
		}
		return sr.ExecuteEntryActions(transition, args...)
	}

	return nil
}

// Exit executes exit actions for this state.
func (sr *StateRepresentation[TState, TTrigger]) Exit(transition Transition[TState, TTrigger]) error {
	if transition.IsReentry() {
		return sr.ExecuteExitActions(transition)
	}

	if !sr.Includes(transition.Destination) {
		if err := sr.ExecuteExitActions(transition); err != nil {
			return err
		}
		if sr.superstate != nil {
			return sr.superstate.Exit(transition)
		}
	}

	return nil
}

// ExecuteEntryActions executes all entry actions for this state.
func (sr *StateRepresentation[TState, TTrigger]) ExecuteEntryActions(transition Transition[TState, TTrigger], args ...any) error {
	for _, action := range sr.entryActions {
		if err := action.Execute(transition, args...); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteExitActions executes all exit actions for this state.
func (sr *StateRepresentation[TState, TTrigger]) ExecuteExitActions(transition Transition[TState, TTrigger]) error {
	for _, action := range sr.exitActions {
		if err := action.Execute(transition); err != nil {
			return err
		}
	}
	return nil
}

// Activate executes activation actions for this state and its superstates.
func (sr *StateRepresentation[TState, TTrigger]) Activate() error {
	if sr.superstate != nil {
		if err := sr.superstate.Activate(); err != nil {
			return err
		}
	}

	return sr.ExecuteActivateActions()
}

// Deactivate executes deactivation actions for this state and its superstates.
func (sr *StateRepresentation[TState, TTrigger]) Deactivate() error {
	if err := sr.ExecuteDeactivateActions(); err != nil {
		return err
	}

	if sr.superstate != nil {
		return sr.superstate.Deactivate()
	}

	return nil
}

// ExecuteActivateActions executes all activation actions for this state.
func (sr *StateRepresentation[TState, TTrigger]) ExecuteActivateActions() error {
	for _, action := range sr.activateActions {
		if err := action.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteDeactivateActions executes all deactivation actions for this state.
func (sr *StateRepresentation[TState, TTrigger]) ExecuteDeactivateActions() error {
	for _, action := range sr.deactivateActions {
		if err := action.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// Includes returns true if this state or any of its substates is the specified state.
func (sr *StateRepresentation[TState, TTrigger]) Includes(state TState) bool {
	if sr.state == state {
		return true
	}
	for _, substate := range sr.substates {
		if substate.Includes(state) {
			return true
		}
	}
	return false
}

// IsIncludedIn returns true if this state is the specified state or a substate of it.
func (sr *StateRepresentation[TState, TTrigger]) IsIncludedIn(state TState) bool {
	if sr.state == state {
		return true
	}
	if sr.superstate != nil {
		return sr.superstate.IsIncludedIn(state)
	}
	return false
}

// GetPermittedTriggers returns the triggers that are currently permitted from this state.
func (sr *StateRepresentation[TState, TTrigger]) GetPermittedTriggers(args ...any) []TTrigger {
	result := sr.GetLocalPermittedTriggers(args...)

	if sr.superstate != nil {
		superTriggers := sr.superstate.GetPermittedTriggers(args...)
		for _, trigger := range superTriggers {
			if !containsTrigger(result, trigger) {
				result = append(result, trigger)
			}
		}
	}

	return result
}

// GetLocalPermittedTriggers returns the triggers that are permitted from this state (not including superstates).
func (sr *StateRepresentation[TState, TTrigger]) GetLocalPermittedTriggers(args ...any) []TTrigger {
	var result []TTrigger
	for trigger, behaviours := range sr.triggerBehaviours {
		for _, behaviour := range behaviours {
			if behaviour.GuardConditionsMet(args...) {
				result = append(result, trigger)
				break
			}
		}
	}
	return result
}

// String returns a string representation of this state.
func (sr *StateRepresentation[TState, TTrigger]) String() string {
	return fmt.Sprintf("%v", sr.state)
}

// containsTrigger checks if a trigger is in the slice.
func containsTrigger[TTrigger comparable](triggers []TTrigger, trigger TTrigger) bool {
	for _, t := range triggers {
		if t == trigger {
			return true
		}
	}
	return false
}
