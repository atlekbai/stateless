package stateless

import (
	"fmt"
)

// StateConfiguration provides a fluent interface for configuring state behaviour.
type StateConfiguration[TState, TTrigger comparable] struct {
	representation *StateRepresentation[TState, TTrigger]
	lookup         func(TState) *StateRepresentation[TState, TTrigger]
}

// NewStateConfiguration creates a new state configuration.
func NewStateConfiguration[TState, TTrigger comparable](
	representation *StateRepresentation[TState, TTrigger],
	lookup func(TState) *StateRepresentation[TState, TTrigger],
) *StateConfiguration[TState, TTrigger] {
	return &StateConfiguration[TState, TTrigger]{
		representation: representation,
		lookup:         lookup,
	}
}

// State returns the state being configured.
func (sc *StateConfiguration[TState, TTrigger]) State() TState {
	return sc.representation.UnderlyingState()
}

// Permit configures the state to transition to the specified destination state
// when the specified trigger is fired.
func (sc *StateConfiguration[TState, TTrigger]) Permit(trigger TTrigger, destinationState TState) *StateConfiguration[TState, TTrigger] {
	sc.enforceNotIdentityTransition(destinationState)
	sc.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(trigger, destinationState, EmptyTransitionGuard),
	)
	return sc
}

// PermitIf configures the state to transition to the specified destination state
// when the specified trigger is fired, if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitIf(trigger TTrigger, destinationState TState, guard func() bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	sc.enforceNotIdentityTransition(destinationState)
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(trigger, destinationState, NewTransitionGuard(guard, desc)),
	)
	return sc
}

// PermitIfArgs configures the state to transition to the specified destination state
// when the specified trigger is fired, if the guard condition (which receives args) is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitIfArgs(trigger TTrigger, destinationState TState, guard func(args any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	sc.enforceNotIdentityTransition(destinationState)
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(trigger, destinationState, NewTransitionGuardWithArgs(guard, desc)),
	)
	return sc
}

// PermitReentry configures the state to re-enter itself when the specified trigger is fired.
// Entry and exit actions will be executed.
func (sc *StateConfiguration[TState, TTrigger]) PermitReentry(trigger TTrigger) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(trigger, sc.representation.UnderlyingState(), EmptyTransitionGuard),
	)
	return sc
}

// PermitReentryIf configures the state to re-enter itself when the specified trigger is fired,
// if the guard condition is met. Entry and exit actions will be executed.
func (sc *StateConfiguration[TState, TTrigger]) PermitReentryIf(trigger TTrigger, guard func() bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(trigger, sc.representation.UnderlyingState(), NewTransitionGuard(guard, desc)),
	)
	return sc
}

// Ignore configures the state to ignore the specified trigger.
func (sc *StateConfiguration[TState, TTrigger]) Ignore(trigger TTrigger) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](trigger, EmptyTransitionGuard),
	)
	return sc
}

// IgnoreIf configures the state to ignore the specified trigger if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) IgnoreIf(trigger TTrigger, guard func() bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](trigger, NewTransitionGuard(guard, desc)),
	)
	return sc
}

// PermitDynamic configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired.
func (sc *StateConfiguration[TState, TTrigger]) PermitDynamic(
	trigger TTrigger,
	destinationSelector func() TState,
	possibleDestinations ...DynamicStateInfo,
) *StateConfiguration[TState, TTrigger] {
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(trigger),
			GuardConditions: nil,
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, "", TimingSynchronous),
		PossibleDestinationStates:           possibleDestinations,
	}
	sc.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(trigger, func(args any) TState { return destinationSelector() }, EmptyTransitionGuard, info),
	)
	return sc
}

// PermitDynamicArgs configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired, using the trigger arguments.
func (sc *StateConfiguration[TState, TTrigger]) PermitDynamicArgs(
	trigger TTrigger,
	destinationSelector func(args any) TState,
	possibleDestinations ...DynamicStateInfo,
) *StateConfiguration[TState, TTrigger] {
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(trigger),
			GuardConditions: nil,
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, "", TimingSynchronous),
		PossibleDestinationStates:           possibleDestinations,
	}
	sc.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(trigger, destinationSelector, EmptyTransitionGuard, info),
	)
	return sc
}

// PermitDynamicIf configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired, if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitDynamicIf(
	trigger TTrigger,
	destinationSelector func() TState,
	guard func() bool,
	guardDescription ...string,
) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(trigger),
			GuardConditions: []InvocationInfo{CreateInvocationInfo(guard, desc, TimingSynchronous)},
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, "", TimingSynchronous),
	}
	sc.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(trigger, func(args any) TState { return destinationSelector() }, NewTransitionGuard(guard, desc), info),
	)
	return sc
}

// PermitDynamicArgsIf configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired, using the trigger arguments, if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitDynamicArgsIf(
	trigger TTrigger,
	destinationSelector func(args any) TState,
	guard func(args any) bool,
	guardDescription ...string,
) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(trigger),
			GuardConditions: []InvocationInfo{CreateInvocationInfo(guard, desc, TimingSynchronous)},
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, "", TimingSynchronous),
	}
	sc.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(trigger, destinationSelector, NewTransitionGuardWithArgs(guard, desc), info),
	)
	return sc
}

// InternalTransition configures an internal transition where the state is not exited
// and re-entered, and entry/exit actions are not executed.
func (sc *StateConfiguration[TState, TTrigger]) InternalTransition(
	trigger TTrigger,
	action func(),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, nil, func(t internalTransition[TState, TTrigger]) { action() }, ""),
	)
	return sc
}

// InternalTransitionIf configures an internal transition where the state is not exited
// and re-entered, if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) InternalTransitionIf(
	trigger TTrigger,
	guard func() bool,
	action func(),
	guardDescription ...string,
) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, func(args any) bool { return guard() }, func(t internalTransition[TState, TTrigger]) { action() }, desc),
	)
	return sc
}

// OnEntry configures an action to be executed when entering this state.
// For simple actions that don't need transition info, use this method.
// For access to typed transition args, use OnEntryWithTransition instead.
func (sc *StateConfiguration[TState, TTrigger]) OnEntry(action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviour[TState, TTrigger](
			func(transition internalTransition[TState, TTrigger]) { action() },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFrom configures an action to be executed when entering this state from a specific trigger.
// For simple actions that don't need transition info, use this method.
// For access to typed transition args, use OnEntryFromWithTransition instead.
func (sc *StateConfiguration[TState, TTrigger]) OnEntryFrom(trigger TTrigger, action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom[TState, TTrigger](
			trigger,
			func(transition internalTransition[TState, TTrigger]) { action() },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnExit configures an action to be executed when exiting this state.
// For simple actions that don't need transition info, use this method.
// For access to typed transition args, use OnExitWithTransition instead.
func (sc *StateConfiguration[TState, TTrigger]) OnExit(action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddExitAction(
		NewSyncExitActionBehaviour[TState, TTrigger](
			func(transition internalTransition[TState, TTrigger]) { action() },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnActivate configures an action to be executed when the state machine is activated
// and this state is the current state.
func (sc *StateConfiguration[TState, TTrigger]) OnActivate(action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddActivateAction(
		NewSyncActivateActionBehaviour(sc.representation.UnderlyingState(), action, CreateInvocationInfo(action, "", TimingSynchronous)),
	)
	return sc
}

// OnDeactivate configures an action to be executed when the state machine is deactivated
// and this state is the current state.
func (sc *StateConfiguration[TState, TTrigger]) OnDeactivate(action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddDeactivateAction(
		NewSyncDeactivateActionBehaviour(sc.representation.UnderlyingState(), action, CreateInvocationInfo(action, "", TimingSynchronous)),
	)
	return sc
}

// SubstateOf sets the superstate of this state.
func (sc *StateConfiguration[TState, TTrigger]) SubstateOf(superstate TState) *StateConfiguration[TState, TTrigger] {
	superstateRep := sc.lookup(superstate)
	if superstateRep == nil {
		panic(fmt.Sprintf("superstate %v not found", superstate))
	}

	// Check for circular references
	if superstateRep.IsIncludedIn(sc.representation.UnderlyingState()) {
		panic(fmt.Sprintf("circular superstate relationship detected: %v -> %v", sc.representation.UnderlyingState(), superstate))
	}

	sc.representation.SetSuperstate(superstateRep)
	superstateRep.AddSubstate(sc.representation)
	return sc
}

// InitialTransition sets the initial transition for this state (used with substates).
// The destination state must be a substate of this state.
func (sc *StateConfiguration[TState, TTrigger]) InitialTransition(destinationState TState) *StateConfiguration[TState, TTrigger] {
	if sc.representation.UnderlyingState() == destinationState {
		panic(fmt.Sprintf("initial transition to self is not allowed: state '%v'", destinationState))
	}
	if sc.representation.HasInitialTransition() {
		panic(fmt.Sprintf("state '%v' already has an initial transition defined", sc.representation.UnderlyingState()))
	}
	sc.representation.SetInitialTransition(destinationState)
	return sc
}

// enforceNotIdentityTransition ensures that a transition is not to the same state.
func (sc *StateConfiguration[TState, TTrigger]) enforceNotIdentityTransition(destinationState TState) {
	if sc.representation.UnderlyingState() == destinationState {
		panic(fmt.Sprintf("permit() requires that the destination state is not equal to the source state. To accept a trigger without changing state, use either Ignore() or PermitReentry()"))
	}
}

// OnEntryWithTransition is a generic function that configures a typed entry action.
// Use this for type-safe access to transition args.
//
// Example:
//
//	type AssignArgs struct { Assignee string }
//	OnEntryWithTransition(sm.Configure(StateB), func(t Transition[State, Trigger, AssignArgs]) {
//	    fmt.Printf("Assigned to %s\n", t.Args.Assignee)
//	})
func OnEntryWithTransition[TState, TTrigger comparable, TArgs any](
	sc *StateConfiguration[TState, TTrigger],
	action func(Transition[TState, TTrigger, TArgs]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviour[TState, TTrigger](
			func(transition internalTransition[TState, TTrigger]) {
				typedTransition := toTypedTransition[TState, TTrigger, TArgs](transition)
				action(typedTransition)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFromWithTransition is a generic function that configures a typed entry action for a specific trigger.
//
// Example:
//
//	type AssignArgs struct { Assignee string }
//	OnEntryFromWithTransition(sm.Configure(StateB), TriggerAssign, func(t Transition[State, Trigger, AssignArgs]) {
//	    fmt.Printf("Assigned to %s\n", t.Args.Assignee)
//	})
func OnEntryFromWithTransition[TState, TTrigger comparable, TArgs any](
	sc *StateConfiguration[TState, TTrigger],
	trigger TTrigger,
	action func(Transition[TState, TTrigger, TArgs]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom[TState, TTrigger](
			trigger,
			func(transition internalTransition[TState, TTrigger]) {
				typedTransition := toTypedTransition[TState, TTrigger, TArgs](transition)
				action(typedTransition)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnExitWithTransition is a generic function that configures a typed exit action.
//
// Example:
//
//	OnExitWithTransition(sm.Configure(StateB), func(t Transition[State, Trigger, NoArgs]) {
//	    fmt.Printf("Exiting from %v\n", t.Source)
//	})
func OnExitWithTransition[TState, TTrigger comparable, TArgs any](
	sc *StateConfiguration[TState, TTrigger],
	action func(Transition[TState, TTrigger, TArgs]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddExitAction(
		NewSyncExitActionBehaviour[TState, TTrigger](
			func(transition internalTransition[TState, TTrigger]) {
				typedTransition := toTypedTransition[TState, TTrigger, TArgs](transition)
				action(typedTransition)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// InternalTransitionWithTransition configures an internal transition with typed transition info.
// This is a package-level function because Go methods cannot have additional type parameters.
func InternalTransitionWithTransition[TState, TTrigger comparable, TArgs any](
	sc *StateConfiguration[TState, TTrigger],
	trigger TTrigger,
	action func(Transition[TState, TTrigger, TArgs]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, nil, func(t internalTransition[TState, TTrigger]) {
			typedTransition := toTypedTransition[TState, TTrigger, TArgs](t)
			action(typedTransition)
		}, ""),
	)
	return sc
}
