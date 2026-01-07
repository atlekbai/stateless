package stateless

import (
	"fmt"
)

// StateConfiguration provides a fluent interface for configuring state behaviour.
type StateConfiguration[TState, TTrigger comparable] struct {
	representation         *StateRepresentation[TState, TTrigger]
	lookup                 func(TState) *StateRepresentation[TState, TTrigger]
	triggerConfigurations  map[TTrigger]*TriggerWithParameters[TTrigger]
}

// NewStateConfiguration creates a new state configuration.
func NewStateConfiguration[TState, TTrigger comparable](
	representation *StateRepresentation[TState, TTrigger],
	lookup func(TState) *StateRepresentation[TState, TTrigger],
	triggerConfigurations map[TTrigger]*TriggerWithParameters[TTrigger],
) *StateConfiguration[TState, TTrigger] {
	return &StateConfiguration[TState, TTrigger]{
		representation:        representation,
		lookup:                lookup,
		triggerConfigurations: triggerConfigurations,
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

// PermitIfWithArgs configures the state to transition to the specified destination state
// when the specified trigger is fired, if the guard condition (which receives arguments) is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitIfWithArgs(trigger TTrigger, destinationState TState, guard func(args ...any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
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

// PermitReentryIfWithArgs configures the state to re-enter itself when the specified trigger is fired,
// if the guard condition (which receives arguments) is met. Entry and exit actions will be executed.
func (sc *StateConfiguration[TState, TTrigger]) PermitReentryIfWithArgs(trigger TTrigger, guard func(args ...any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewReentryTriggerBehaviour(trigger, sc.representation.UnderlyingState(), NewTransitionGuardWithArgs(guard, desc)),
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

// IgnoreIfWithArgs configures the state to ignore the specified trigger if the guard condition
// (which receives arguments) is met.
func (sc *StateConfiguration[TState, TTrigger]) IgnoreIfWithArgs(trigger TTrigger, guard func(args ...any) bool, guardDescription ...string) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewIgnoredTriggerBehaviour[TState](trigger, NewTransitionGuardWithArgs(guard, desc)),
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
		NewDynamicTriggerBehaviour(trigger, func(args ...any) TState { return destinationSelector() }, EmptyTransitionGuard, info),
	)
	return sc
}

// PermitDynamicWithArgs configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired, using the trigger arguments.
func (sc *StateConfiguration[TState, TTrigger]) PermitDynamicWithArgs(
	trigger TTrigger,
	destinationSelector func(args ...any) TState,
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
		NewDynamicTriggerBehaviour(trigger, func(args ...any) TState { return destinationSelector() }, NewTransitionGuard(guard, desc), info),
	)
	return sc
}

// PermitDynamicIfWithArgs configures the state to transition to a dynamically determined destination state
// when the specified trigger is fired, if the guard condition (which receives arguments) is met.
func (sc *StateConfiguration[TState, TTrigger]) PermitDynamicIfWithArgs(
	trigger TTrigger,
	destinationSelector func(args ...any) TState,
	guard func(args ...any) bool,
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
	action func(transition Transition[TState, TTrigger], args ...any),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, nil, action, ""),
	)
	return sc
}

// InternalTransitionIf configures an internal transition where the state is not exited
// and re-entered, if the guard condition is met.
func (sc *StateConfiguration[TState, TTrigger]) InternalTransitionIf(
	trigger TTrigger,
	guard func(args ...any) bool,
	action func(transition Transition[TState, TTrigger], args ...any),
	guardDescription ...string,
) *StateConfiguration[TState, TTrigger] {
	desc := ""
	if len(guardDescription) > 0 {
		desc = guardDescription[0]
	}
	sc.representation.AddTriggerBehaviour(
		NewSyncInternalTriggerBehaviour(trigger, guard, action, desc),
	)
	return sc
}

// OnEntry configures an action to be executed when entering this state.
func (sc *StateConfiguration[TState, TTrigger]) OnEntry(action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviour[TState, TTrigger](
			func(transition Transition[TState, TTrigger], args ...any) { action() },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryWithTransition configures an action to be executed when entering this state,
// receiving the transition that caused the entry.
func (sc *StateConfiguration[TState, TTrigger]) OnEntryWithTransition(action func(transition Transition[TState, TTrigger])) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviour[TState, TTrigger](
			func(transition Transition[TState, TTrigger], args ...any) { action(transition) },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryWithArgs configures an action to be executed when entering this state,
// receiving the trigger arguments.
func (sc *StateConfiguration[TState, TTrigger]) OnEntryWithArgs(action func(transition Transition[TState, TTrigger], args ...any)) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviour[TState, TTrigger](action, CreateInvocationInfo(action, "", TimingSynchronous)),
	)
	return sc
}

// OnEntryFrom configures an action to be executed when entering this state from a specific trigger.
func (sc *StateConfiguration[TState, TTrigger]) OnEntryFrom(trigger TTrigger, action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger,
			func(transition Transition[TState, TTrigger], args ...any) { action() },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFromWithTransition configures an action to be executed when entering this state
// from a specific trigger, receiving the transition.
func (sc *StateConfiguration[TState, TTrigger]) OnEntryFromWithTransition(trigger TTrigger, action func(transition Transition[TState, TTrigger])) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger,
			func(transition Transition[TState, TTrigger], args ...any) { action(transition) },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFromWithArgs configures an action to be executed when entering this state
// from a specific trigger, receiving the trigger arguments.
func (sc *StateConfiguration[TState, TTrigger]) OnEntryFromWithArgs(trigger TTrigger, action func(transition Transition[TState, TTrigger], args ...any)) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(trigger, action, CreateInvocationInfo(action, "", TimingSynchronous)),
	)
	return sc
}

// OnExit configures an action to be executed when exiting this state.
func (sc *StateConfiguration[TState, TTrigger]) OnExit(action func()) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddExitAction(
		NewSyncExitActionBehaviour[TState, TTrigger](
			func(transition Transition[TState, TTrigger]) { action() },
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnExitWithTransition configures an action to be executed when exiting this state,
// receiving the transition that caused the exit.
func (sc *StateConfiguration[TState, TTrigger]) OnExitWithTransition(action func(transition Transition[TState, TTrigger])) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddExitAction(
		NewSyncExitActionBehaviour[TState, TTrigger](action, CreateInvocationInfo(action, "", TimingSynchronous)),
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
func (sc *StateConfiguration[TState, TTrigger]) InitialTransition(destinationState TState) *StateConfiguration[TState, TTrigger] {
	sc.representation.SetInitialTransition(destinationState)
	return sc
}

// enforceNotIdentityTransition ensures that a transition is not to the same state.
func (sc *StateConfiguration[TState, TTrigger]) enforceNotIdentityTransition(destinationState TState) {
	if sc.representation.UnderlyingState() == destinationState {
		panic(fmt.Sprintf("permit() requires that the destination state is not equal to the source state. To accept a trigger without changing state, use either Ignore() or PermitReentry()"))
	}
}

// OnEntryFrom1 configures an action to be executed when entering this state
// from a specific parameterized trigger, receiving the typed argument.
// This is a standalone function because Go methods cannot introduce new type parameters.
func OnEntryFrom1[TState, TTrigger comparable, TArg0 any](
	sc *StateConfiguration[TState, TTrigger],
	trigger *TriggerWithParameters1[TTrigger, TArg0],
	action func(arg0 TArg0),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger.Trigger(),
			func(transition Transition[TState, TTrigger], args ...any) {
				var arg0 TArg0
				if len(args) > 0 {
					arg0, _ = args[0].(TArg0)
				}
				action(arg0)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFrom1WithTransition configures an action to be executed when entering this state
// from a specific parameterized trigger, receiving the typed argument and transition.
func OnEntryFrom1WithTransition[TState, TTrigger comparable, TArg0 any](
	sc *StateConfiguration[TState, TTrigger],
	trigger *TriggerWithParameters1[TTrigger, TArg0],
	action func(arg0 TArg0, transition Transition[TState, TTrigger]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger.Trigger(),
			func(transition Transition[TState, TTrigger], args ...any) {
				var arg0 TArg0
				if len(args) > 0 {
					arg0, _ = args[0].(TArg0)
				}
				action(arg0, transition)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFrom2 configures an action to be executed when entering this state
// from a specific parameterized trigger, receiving two typed arguments.
func OnEntryFrom2[TState, TTrigger comparable, TArg0, TArg1 any](
	sc *StateConfiguration[TState, TTrigger],
	trigger *TriggerWithParameters2[TTrigger, TArg0, TArg1],
	action func(arg0 TArg0, arg1 TArg1),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger.Trigger(),
			func(transition Transition[TState, TTrigger], args ...any) {
				var arg0 TArg0
				var arg1 TArg1
				if len(args) > 0 {
					arg0, _ = args[0].(TArg0)
				}
				if len(args) > 1 {
					arg1, _ = args[1].(TArg1)
				}
				action(arg0, arg1)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFrom2WithTransition configures an action to be executed when entering this state
// from a specific parameterized trigger, receiving two typed arguments and transition.
func OnEntryFrom2WithTransition[TState, TTrigger comparable, TArg0, TArg1 any](
	sc *StateConfiguration[TState, TTrigger],
	trigger *TriggerWithParameters2[TTrigger, TArg0, TArg1],
	action func(arg0 TArg0, arg1 TArg1, transition Transition[TState, TTrigger]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger.Trigger(),
			func(transition Transition[TState, TTrigger], args ...any) {
				var arg0 TArg0
				var arg1 TArg1
				if len(args) > 0 {
					arg0, _ = args[0].(TArg0)
				}
				if len(args) > 1 {
					arg1, _ = args[1].(TArg1)
				}
				action(arg0, arg1, transition)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFrom3 configures an action to be executed when entering this state
// from a specific parameterized trigger, receiving three typed arguments.
func OnEntryFrom3[TState, TTrigger comparable, TArg0, TArg1, TArg2 any](
	sc *StateConfiguration[TState, TTrigger],
	trigger *TriggerWithParameters3[TTrigger, TArg0, TArg1, TArg2],
	action func(arg0 TArg0, arg1 TArg1, arg2 TArg2),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger.Trigger(),
			func(transition Transition[TState, TTrigger], args ...any) {
				var arg0 TArg0
				var arg1 TArg1
				var arg2 TArg2
				if len(args) > 0 {
					arg0, _ = args[0].(TArg0)
				}
				if len(args) > 1 {
					arg1, _ = args[1].(TArg1)
				}
				if len(args) > 2 {
					arg2, _ = args[2].(TArg2)
				}
				action(arg0, arg1, arg2)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}

// OnEntryFrom3WithTransition configures an action to be executed when entering this state
// from a specific parameterized trigger, receiving three typed arguments and transition.
func OnEntryFrom3WithTransition[TState, TTrigger comparable, TArg0, TArg1, TArg2 any](
	sc *StateConfiguration[TState, TTrigger],
	trigger *TriggerWithParameters3[TTrigger, TArg0, TArg1, TArg2],
	action func(arg0 TArg0, arg1 TArg1, arg2 TArg2, transition Transition[TState, TTrigger]),
) *StateConfiguration[TState, TTrigger] {
	sc.representation.AddEntryAction(
		NewSyncEntryActionBehaviourFrom(
			trigger.Trigger(),
			func(transition Transition[TState, TTrigger], args ...any) {
				var arg0 TArg0
				var arg1 TArg1
				var arg2 TArg2
				if len(args) > 0 {
					arg0, _ = args[0].(TArg0)
				}
				if len(args) > 1 {
					arg1, _ = args[1].(TArg1)
				}
				if len(args) > 2 {
					arg2, _ = args[2].(TArg2)
				}
				action(arg0, arg1, arg2, transition)
			},
			CreateInvocationInfo(action, "", TimingSynchronous),
		),
	)
	return sc
}
