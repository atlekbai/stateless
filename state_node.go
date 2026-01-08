package stateless

import (
	"context"
	"fmt"
)

// StateNode provides a fluent interface for configuring state behaviour.
type StateNode[TState, TTrigger comparable] struct {
	representation *StateRepresentation[TState, TTrigger]
	lookup         func(TState) *StateRepresentation[TState, TTrigger]
}

// NewStateNode creates a new state configuration.
func NewStateNode[TState, TTrigger comparable](
	representation *StateRepresentation[TState, TTrigger],
	lookup func(TState) *StateRepresentation[TState, TTrigger],
) *StateNode[TState, TTrigger] {
	return &StateNode[TState, TTrigger]{
		representation: representation,
		lookup:         lookup,
	}
}

// State returns the state being configured.
func (sn *StateNode[TState, TTrigger]) State() TState {
	return sn.representation.UnderlyingState()
}

// Permit configures the state to transition to the specified destination state
// when the specified trigger is fired.
func (sn *StateNode[TState, TTrigger]) Permit(tr TTrigger, dst TState) *StateNode[TState, TTrigger] {
	sn.enforceNotIdentityTransition(dst)
	sn.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(tr, dst, EmptyTransitionGuard),
	)
	return sn
}

// PermitIf configures the state to transition to the specified destination state
// when the specified trigger is fired, if the guard condition is met.
// The guard returns nil if the condition is met, or an error describing why it failed.
func (sn *StateNode[TState, TTrigger]) PermitIf(tr TTrigger, dst TState, gf GuardFunc) *StateNode[TState, TTrigger] {
	sn.enforceNotIdentityTransition(dst)
	sn.representation.AddTriggerBehaviour(
		NewTransitioningTriggerBehaviour(tr, dst, NewTransitionGuard(gf)),
	)
	return sn
}

// PermitReentry configures the state to re-enter itself when the specified trigger is fired.
// Entry and exit actions will be executed.
func (sn *StateNode[TState, TTrigger]) PermitReentry(tr TTrigger) *StateNode[TState, TTrigger] {
	sn.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(tr, sn.representation.UnderlyingState(), EmptyTransitionGuard),
	)
	return sn
}

// PermitReentryIf configures the state to re-enter itself when the specified trigger is fired,
// if the guard condition is met. Entry and exit actions will be executed.
// The guard returns nil if the condition is met, or an error describing why it failed.
func (sn *StateNode[TState, TTrigger]) PermitReentryIf(tr TTrigger, gf GuardFunc) *StateNode[TState, TTrigger] {
	sn.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(
			tr,
			sn.representation.UnderlyingState(),
			NewTransitionGuard(gf),
		),
	)
	return sn
}

// Ignore configures the state to ignore the specified trigger.
func (sn *StateNode[TState, TTrigger]) Ignore(tr TTrigger) *StateNode[TState, TTrigger] {
	sn.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](tr, EmptyTransitionGuard),
	)
	return sn
}

// IgnoreIf configures the state to ignore the specified trigger if the guard condition is met.
// The guard returns nil if the condition is met, or an error describing why it failed.
func (sn *StateNode[TState, TTrigger]) IgnoreIf(tr TTrigger, gf GuardFunc) *StateNode[TState, TTrigger] {
	sn.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](tr, NewTransitionGuard(gf)),
	)
	return sn
}

// PermitDynamic configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired. The destination selector receives the trigger arguments.
// If you don't need args, use func(_ any) TState { return targetState }.
func (sn *StateNode[TState, TTrigger]) PermitDynamic(
	tr TTrigger,
	ss StateSelector[TState],
	possibleDestinations ...DynamicStateInfo,
) *StateNode[TState, TTrigger] {
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(tr),
			GuardConditions: nil,
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(ss, ""),
		PossibleDestinationStates:           possibleDestinations,
	}
	sn.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(tr, ss, EmptyTransitionGuard, info),
	)
	return sn
}

// PermitDynamicIf configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired, if the guard condition is met.
// Both selector and guard receive the trigger arguments.
// The guard returns nil if the condition is met, or an error describing why it failed.
func (sn *StateNode[TState, TTrigger]) PermitDynamicIf(
	tr TTrigger,
	ss StateSelector[TState],
	gf GuardFunc,
) *StateNode[TState, TTrigger] {
	info := DynamicTransitionInfo{
		transitionInfoBase: transitionInfoBase{
			Trigger:         NewTriggerInfo(tr),
			GuardConditions: []InvocationInfo{CreateInvocationInfo(gf, "")},
		},
		DestinationStateSelectorDescription: CreateInvocationInfo(ss, ""),
	}
	sn.representation.AddTriggerBehaviour(
		NewDynamicTriggerBehaviour(tr, ss, NewTransitionGuard(gf), info),
	)
	return sn
}

// InternalTransition configures an internal transition where the state is not exited
// and re-entered, and entry/exit actions are not executed.
func (sn *StateNode[TState, TTrigger]) InternalTransition(
	tr TTrigger,
	act TransitionAction[TState, TTrigger],
) *StateNode[TState, TTrigger] {
	sn.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(tr, EmptyTransitionGuard, act),
	)
	return sn
}

// InternalTransitionIf configures an internal transition where the state is not exited
// and re-entered, if the guard condition is met.
// The guard returns nil if the condition is met, or an error describing why it failed.
func (sn *StateNode[TState, TTrigger]) InternalTransitionIf(
	tr TTrigger,
	gf GuardFunc,
	act TransitionAction[TState, TTrigger],
) *StateNode[TState, TTrigger] {
	sn.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(tr, NewTransitionGuard(gf), act),
	)
	return sn
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
func (sn *StateNode[TState, TTrigger]) OnEntry(act TransitionAction[TState, TTrigger]) *StateNode[TState, TTrigger] {
	sn.representation.AddEntryAction(
		NewSyncEntryActionBehaviour(act, CreateInvocationInfo(act, "")),
	)
	return sn
}

// OnExit configures an action to be executed when exiting this state.
// The action receives the transition information including source, destination, trigger, and args.
func (sn *StateNode[TState, TTrigger]) OnExit(act TransitionAction[TState, TTrigger]) *StateNode[TState, TTrigger] {
	sn.representation.AddExitAction(
		NewSyncExitActionBehaviour(act, CreateInvocationInfo(act, "")),
	)
	return sn
}

// OnActivate configures an action to be executed when the state machine is activated
// and this state is the current state.
func (sn *StateNode[TState, TTrigger]) OnActivate(act func(ctx context.Context) error) *StateNode[TState, TTrigger] {
	sn.representation.AddActivateAction(
		NewSyncActivateActionBehaviour[TState](act, CreateInvocationInfo(act, "")),
	)
	return sn
}

// OnDeactivate configures an action to be executed when the state machine is deactivated
// and this state is the current state.
func (sn *StateNode[TState, TTrigger]) OnDeactivate(act func(ctx context.Context) error) *StateNode[TState, TTrigger] {
	sn.representation.AddDeactivateAction(
		NewSyncDeactivateActionBehaviour[TState](act, CreateInvocationInfo(act, "")),
	)
	return sn
}

// SubstateOf sets the superstate of this state.
func (sn *StateNode[TState, TTrigger]) SubstateOf(superstate TState) *StateNode[TState, TTrigger] {
	superstateRep := sn.lookup(superstate)
	if superstateRep == nil {
		panic(fmt.Sprintf("superstate %v not found", superstate))
	}

	// Check for circular references
	if superstateRep.IsIncludedIn(sn.representation.UnderlyingState()) {
		panic(fmt.Sprintf(
			"circular superstate relationship detected: %v -> %v",
			sn.representation.UnderlyingState(),
			superstate,
		))
	}

	sn.representation.SetSuperstate(superstateRep)
	superstateRep.AddSubstate(sn.representation)
	return sn
}

// InitialTransition sets the initial transition for this state (used with substates).
// The destination state must be a substate of this state.
func (sn *StateNode[TState, TTrigger]) InitialTransition(dst TState) *StateNode[TState, TTrigger] {
	if sn.representation.UnderlyingState() == dst {
		panic(fmt.Sprintf("initial transition to self is not allowed: state '%v'", dst))
	}
	if sn.representation.HasInitialTransition() {
		panic(fmt.Sprintf("state '%v' already has an initial transition defined", sn.representation.UnderlyingState()))
	}
	sn.representation.SetInitialTransition(dst)
	return sn
}

// enforceNotIdentityTransition ensures that a transition is not to the same state.
func (sn *StateNode[TState, TTrigger]) enforceNotIdentityTransition(dst TState) {
	if sn.representation.UnderlyingState() == dst {
		panic(
			"permit() requires that the destination state is not equal to the source state. To accept a trigger without changing state, use either Ignore() or PermitReentry()",
		)
	}
}
