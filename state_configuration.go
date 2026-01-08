package stateless

import (
	"context"
	"fmt"
)

// StateConfiguration provides a fluent interface for configuring state behaviour.
type StateConfiguration[TState, TTrigger comparable] struct {
	representation *StateRepresentation[TState, TTrigger]
	lookup         func(TState) *StateRepresentation[TState, TTrigger]
}

// firstOrEmpty returns the first element of the slice or empty string if empty.
func firstOrEmpty(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
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
	sc.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(trigger, destinationState, NewTransitionGuard(guard, firstOrEmpty(guardDescription))),
	)
	return sc
}

// PermitIfArgs configures the state to transition to the specified destination state
// when the specified trigger is fired, if the guard condition (which receives args) is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitIfArgs(trigger TTrigger, destinationState TState, guard func(args any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	sc.enforceNotIdentityTransition(destinationState)
	sc.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(trigger, destinationState, NewTransitionGuardWithArgs(guard, firstOrEmpty(guardDescription))),
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
	sc.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(trigger, sc.representation.UnderlyingState(), NewTransitionGuard(guard, firstOrEmpty(guardDescription))),
	)
	return sc
}

// PermitReentryIfArgs configures the state to re-enter itself when the specified trigger is fired,
// if the guard condition (which receives args) is met. Entry and exit actions will be executed.
func (sc *StateConfiguration[TState, TTrigger]) PermitReentryIfArgs(trigger TTrigger, guard func(args any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(trigger, sc.representation.UnderlyingState(), NewTransitionGuardWithArgs(guard, firstOrEmpty(guardDescription))),
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
	sc.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](trigger, NewTransitionGuard(guard, firstOrEmpty(guardDescription))),
	)
	return sc
}

// IgnoreIfArgs configures the state to ignore the specified trigger if the guard condition (which receives args) is met.
func (sc *StateConfiguration[TState, TTrigger]) IgnoreIfArgs(trigger TTrigger, guard func(args any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](trigger, NewTransitionGuardWithArgs(guard, firstOrEmpty(guardDescription))),
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
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, ""),
		PossibleDestinationStates:           possibleDestinations,
	}
	sc.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(trigger, func(_ any) TState { return destinationSelector() }, EmptyTransitionGuard, info),
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
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, ""),
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
	desc := firstOrEmpty(guardDescription)
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(trigger),
			GuardConditions: []InvocationInfo{CreateInvocationInfo(guard, desc)},
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, ""),
	}
	sc.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(trigger, func(_ any) TState { return destinationSelector() }, NewTransitionGuard(guard, desc), info),
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
	desc := firstOrEmpty(guardDescription)
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(trigger),
			GuardConditions: []InvocationInfo{CreateInvocationInfo(guard, desc)},
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(destinationSelector, ""),
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
	action func(ctx context.Context, t Transition[TState, TTrigger]) error,
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, EmptyTransitionGuard, action),
	)
	return sc
}

// InternalTransitionIf configures an internal transition where the state is not exited
// and re-entered, if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) InternalTransitionIf(
	trigger TTrigger,
	guard func() bool,
	action func(ctx context.Context, t Transition[TState, TTrigger]) error,
	guardDescription ...string,
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, NewTransitionGuard(guard, firstOrEmpty(guardDescription)), action),
	)
	return sc
}

// InternalTransitionIfArgs configures an internal transition where the state is not exited
// and re-entered, if the guard condition (which receives args) is met.
func (sc *StateConfiguration[TState, TTrigger]) InternalTransitionIfArgs(
	trigger TTrigger,
	guard func(args any) bool,
	action func(ctx context.Context, t Transition[TState, TTrigger]) error,
	guardDescription ...string,
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, NewTransitionGuardWithArgs(guard, firstOrEmpty(guardDescription)), action),
	)
	return sc
}

// OnEntry configures an action to be executed when entering this state.
// The action receives the transition information including source, destination, trigger, and args.
// Use type assertion to access typed arguments:
//
//	OnEntry(func(ctx context.Context, t Transition[State, Trigger]) error {
//	    if args, ok := t.Args.(MyArgs); ok {
//	        // use args
//	    }
//	    return nil
//	})
func (sc *StateConfiguration[TState, TTrigger]) OnEntry(action func(ctx context.Context, t Transition[TState, TTrigger]) error) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviour[TState, TTrigger](action, CreateInvocationInfo(action, "")),
	)
	return sc
}

// OnExit configures an action to be executed when exiting this state.
// The action receives the transition information including source, destination, trigger, and args.
func (sc *StateConfiguration[TState, TTrigger]) OnExit(action func(ctx context.Context, t Transition[TState, TTrigger]) error) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddExitAction(
		NewSyncExitActionBehaviour[TState, TTrigger](action, CreateInvocationInfo(action, "")),
	)
	return sc
}

// OnActivate configures an action to be executed when the state machine is activated
// and this state is the current state.
func (sc *StateConfiguration[TState, TTrigger]) OnActivate(action func(ctx context.Context) error) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddActivateAction(
		NewSyncActivateActionBehaviour[TState](action, CreateInvocationInfo(action, "")),
	)
	return sc
}

// OnDeactivate configures an action to be executed when the state machine is deactivated
// and this state is the current state.
func (sc *StateConfiguration[TState, TTrigger]) OnDeactivate(action func(ctx context.Context) error) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddDeactivateAction(
		NewSyncDeactivateActionBehaviour[TState](action, CreateInvocationInfo(action, "")),
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
