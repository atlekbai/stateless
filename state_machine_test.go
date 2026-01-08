package stateless

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// Test state and trigger types.
type State int
type Trigger int

const (
	StateA State = iota
	StateB
	StateC
	StateD
)

const (
	TriggerX Trigger = iota
	TriggerY
	TriggerZ
)

func (s State) String() string {
	switch s {
	case StateA:
		return "StateA"
	case StateB:
		return "StateB"
	case StateC:
		return "StateC"
	case StateD:
		return "StateD"
	default:
		return "Unknown"
	}
}

func (t Trigger) String() string {
	switch t {
	case TriggerX:
		return "TriggerX"
	case TriggerY:
		return "TriggerY"
	case TriggerZ:
		return "TriggerZ"
	default:
		return "Unknown"
	}
}

// Basic tests

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	if sm.State() != StateA {
		t.Errorf("expected initial state to be StateA, got %v", sm.State())
	}
}

func TestSimpleTransition(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected state to be StateB, got %v", sm.State())
	}
}

func TestMultipleTransitions(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).Permit(TriggerY, StateC)
	sm.Configure(StateC).Permit(TriggerZ, StateA)

	// A -> B
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}

	// B -> C
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}

	// C -> A
	if err := sm.Fire(TriggerZ, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
}

func TestInvalidTransition(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	// TriggerY is not configured for StateA
	err := sm.Fire(TriggerY, nil)
	if err == nil {
		t.Error("expected error for invalid transition")
	}

	var invalidTransitionErr *InvalidTransitionError
	if !errors.As(err, &invalidTransitionErr) {
		t.Errorf("expected InvalidTransitionError, got %T", err)
	}
}

// Guard tests

func TestPermitIf_GuardPasses(t *testing.T) {
	guardCalled := false
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitIf(TriggerX, StateB, func() bool {
		guardCalled = true
		return true
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !guardCalled {
		t.Error("guard was not called")
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

func TestPermitIf_GuardFails(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitIf(TriggerX, StateB, func() bool {
		return false
	}, "test guard")

	err := sm.Fire(TriggerX, nil)
	if err == nil {
		t.Error("expected error when guard fails")
	}

	if sm.State() != StateA {
		t.Errorf("expected state to remain StateA, got %v", sm.State())
	}
}

func TestMultipleGuards(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	// First guard passes -> StateB
	// Second guard fails -> StateC
	sm.Configure(StateA).
		PermitIf(TriggerX, StateB, func() bool { return true }, "guard1").
		PermitIf(TriggerX, StateC, func() bool { return false }, "guard2")

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

// Entry/Exit action tests

func TestOnEntry(t *testing.T) {
	entryCount := 0
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
		entryCount++
		return nil
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if entryCount != 1 {
		t.Errorf("expected entry action to be called once, got %d", entryCount)
	}
}

func TestOnExit(t *testing.T) {
	exitCount := 0
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error {
			exitCount++
			return nil
		})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if exitCount != 1 {
		t.Errorf("expected exit action to be called once, got %d", exitCount)
	}
}

func TestOnEntryWithTransition(t *testing.T) {
	var receivedTransition Transition[State, Trigger]
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).OnEntry(func(ctx context.Context, transition Transition[State, Trigger]) error {
		receivedTransition = transition
		return nil
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if receivedTransition.Source != StateA {
		t.Errorf("expected source StateA, got %v", receivedTransition.Source)
	}
	if receivedTransition.Destination != StateB {
		t.Errorf("expected destination StateB, got %v", receivedTransition.Destination)
	}
	if receivedTransition.Trigger != TriggerX {
		t.Errorf("expected trigger TriggerX, got %v", receivedTransition.Trigger)
	}
}

func TestOnEntryCheckTrigger(t *testing.T) {
	// This test shows how to check the trigger in OnEntry
	// (replaces the old OnEntryFrom functionality)
	entryFromXCount := 0
	entryFromYCount := 0

	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		Permit(TriggerY, StateC).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
			if tr.Trigger == TriggerX {
				entryFromXCount++
			} else if tr.Trigger == TriggerY {
				entryFromYCount++
			}
			return nil
		})

	sm.Configure(StateC).Permit(TriggerY, StateB)

	// Fire TriggerX: A -> B
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if entryFromXCount != 1 {
		t.Errorf("expected entryFromXCount to be 1, got %d", entryFromXCount)
	}
	if entryFromYCount != 0 {
		t.Errorf("expected entryFromYCount to be 0, got %d", entryFromYCount)
	}

	// Fire TriggerY: B -> C
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Fire TriggerY: C -> B (should trigger OnEntry with TriggerY)
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if entryFromYCount != 1 {
		t.Errorf("expected entryFromYCount to be 1, got %d", entryFromYCount)
	}
}

// Reentry tests

func TestPermitReentry(t *testing.T) {
	entryCount := 0
	exitCount := 0

	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		PermitReentry(TriggerX).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { entryCount++; return nil }).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { exitCount++; return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if sm.State() != StateA {
		t.Errorf("expected state to remain StateA, got %v", sm.State())
	}
	if entryCount != 1 {
		t.Errorf("expected entry action to be called once, got %d", entryCount)
	}
	if exitCount != 1 {
		t.Errorf("expected exit action to be called once, got %d", exitCount)
	}
}

// Ignore tests

func TestIgnore(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		Ignore(TriggerY)

	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error when firing ignored trigger: %v", err)
	}

	if sm.State() != StateA {
		t.Errorf("expected state to remain StateA, got %v", sm.State())
	}
}

func TestIgnoreIf(t *testing.T) {
	shouldIgnore := true
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		IgnoreIf(TriggerX, func() bool { return shouldIgnore })

	// Should be ignored
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error when trigger should be ignored: %v", err)
	}
	if sm.State() != StateA {
		t.Errorf("expected state to remain StateA, got %v", sm.State())
	}

	// Now it shouldn't be ignored (but no transition defined)
	shouldIgnore = false
	err := sm.Fire(TriggerX, nil)
	if err == nil {
		t.Error("expected error when trigger is not ignored and no transition defined")
	}
}

// Internal transition tests

func TestInternalTransition(t *testing.T) {
	actionCount := 0
	entryCount := 0
	exitCount := 0

	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error {
			actionCount++
			return nil
		}).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { entryCount++; return nil }).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { exitCount++; return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if actionCount != 1 {
		t.Errorf("expected action to be called once, got %d", actionCount)
	}
	if entryCount != 0 {
		t.Errorf("expected entry action not to be called, got %d", entryCount)
	}
	if exitCount != 0 {
		t.Errorf("expected exit action not to be called, got %d", exitCount)
	}
	if sm.State() != StateA {
		t.Errorf("expected state to remain StateA, got %v", sm.State())
	}
}

// Internal transition tests (ported from .NET Stateless)

func TestInternalTransition_StayInSameStateOneState(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil })

	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateA {
		t.Errorf("expected StateA after fire, got %v", sm.State())
	}
}

func TestInternalTransition_StayInSameStateTwoStates(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil }).
		Permit(TriggerY, StateB)

	sm.Configure(StateB).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil }).
		Permit(TriggerY, StateA)

	// This should not cause any state changes
	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateA {
		t.Errorf("expected StateA after TriggerX, got %v", sm.State())
	}

	// Change state to B
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// This should also not cause any state changes
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB after TriggerX, got %v", sm.State())
	}
}

func TestInternalTransition_StayInSameSubStateTransitionInSuperstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil }).
		InternalTransition(TriggerY, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil })

	sm.Configure(StateB).
		SubstateOf(StateA)

	// This should not cause any state changes
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB after TriggerX, got %v", sm.State())
	}
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB after TriggerY, got %v", sm.State())
	}
}

func TestInternalTransition_StayInSameSubStateTransitionInSubstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA)

	sm.Configure(StateB).
		SubstateOf(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil }).
		InternalTransition(TriggerY, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil })

	// This should not cause any state changes
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB after TriggerX, got %v", sm.State())
	}
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB after TriggerY, got %v", sm.State())
	}
}

func TestInternalTransitionIf_ShouldBeReflectedInPermittedTriggers(t *testing.T) {
	isPermitted := true
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		InternalTransitionIf(TriggerX, func() bool { return isPermitted }, func(ctx context.Context, tr Transition[State, Trigger]) error { return nil })

	triggers := sm.GetPermittedTriggers(nil)
	if len(triggers) != 1 {
		t.Errorf("expected 1 permitted trigger, got %d", len(triggers))
	}

	isPermitted = false
	triggers = sm.GetPermittedTriggers(nil)
	if len(triggers) != 0 {
		t.Errorf("expected 0 permitted triggers, got %d", len(triggers))
	}
}

func TestInternalTransition_HandledOnlyOnceInSuper(t *testing.T) {
	handledIn := StateC

	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { handledIn = StateA; return nil })

	sm.Configure(StateB).
		SubstateOf(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { handledIn = StateB; return nil })

	// The state machine is in state A. It should only be handled in State A
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handledIn != StateA {
		t.Errorf("expected handledIn to be StateA, got %v", handledIn)
	}
}

func TestInternalTransition_HandledOnlyOnceInSub(t *testing.T) {
	handledIn := StateC

	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { handledIn = StateA; return nil })

	sm.Configure(StateB).
		SubstateOf(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { handledIn = StateB; return nil })

	// The state machine is in state B. It should only be handled in State B
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handledIn != StateB {
		t.Errorf("expected handledIn to be StateB, got %v", handledIn)
	}
}

func TestInternalTransition_OnlyOneHandlerExecuted(t *testing.T) {
	handled := 0

	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr Transition[State, Trigger]) error { handled++; return nil }).
		InternalTransition(TriggerY, func(ctx context.Context, tr Transition[State, Trigger]) error { handled++; return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handled != 1 {
		t.Errorf("expected handled to be 1, got %d", handled)
	}

	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handled != 2 {
		t.Errorf("expected handled to be 2, got %d", handled)
	}
}

func TestInternalTransitionWithTypedArgs(t *testing.T) {
	type InternalArgs struct {
		Value int
	}

	sm := NewStateMachine[State, Trigger](StateA)
	var receivedValue int

	sm.Configure(StateA).InternalTransition(TriggerX,
		func(ctx context.Context, trans Transition[State, Trigger]) error {
			if args, ok := trans.Args.(InternalArgs); ok {
				receivedValue = args.Value
			}
			return nil
		},
	)

	if err := sm.Fire(TriggerX, InternalArgs{Value: 42}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedValue != 42 {
		t.Errorf("expected receivedValue to be 42, got %d", receivedValue)
	}
	if sm.State() != StateA {
		t.Errorf("expected state to remain StateA, got %v", sm.State())
	}
}

// Dynamic transition tests

func TestPermitDynamic(t *testing.T) {
	destState := StateB
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitDynamic(TriggerX, func() State {
		return destState
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}

	// Reset and try with different destination
	sm = NewStateMachine[State, Trigger](StateA)
	destState = StateC
	sm.Configure(StateA).PermitDynamic(TriggerX, func() State {
		return destState
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestPermitDynamicWithArgs(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitDynamicArgs(TriggerX, func(args any) State {
		if state, ok := args.(State); ok {
			return state
		}
		return StateB
	})

	if err := sm.Fire(TriggerX, StateC); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

// Hierarchical state tests

func TestSubstateOf(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	// StateB is a superstate, StateC is a substate
	sm.Configure(StateB).Permit(TriggerX, StateA)
	sm.Configure(StateC).SubstateOf(StateB)
	sm.Configure(StateA).Permit(TriggerY, StateC)

	// Go to StateC
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}

	// StateC should inherit TriggerX from StateB
	if !sm.CanFire(TriggerX, nil) {
		t.Error("expected TriggerX to be firable from StateC (inherited from StateB)")
	}

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
}

func TestIsInState_WithSubstates(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateC)
	sm.Configure(StateB)
	sm.Configure(StateC).SubstateOf(StateB)

	if !sm.IsInState(StateC) {
		t.Error("expected IsInState(StateC) to be true")
	}
	if !sm.IsInState(StateB) {
		t.Error("expected IsInState(StateB) to be true (StateC is substate of StateB)")
	}
	if sm.IsInState(StateA) {
		t.Error("expected IsInState(StateA) to be false")
	}
}

// Activation/Deactivation tests

func TestActivateDeactivate(t *testing.T) {
	activateCount := 0
	deactivateCount := 0

	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		OnActivate(func(ctx context.Context) error { activateCount++; return nil }).
		OnDeactivate(func(ctx context.Context) error { deactivateCount++; return nil })

	if err := sm.Activate(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if activateCount != 1 {
		t.Errorf("expected activate action to be called once, got %d", activateCount)
	}

	// Calling activate again should be idempotent
	if err := sm.Activate(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if activateCount != 1 {
		t.Errorf("expected activate action to still be 1, got %d", activateCount)
	}

	if err := sm.Deactivate(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if deactivateCount != 1 {
		t.Errorf("expected deactivate action to be called once, got %d", deactivateCount)
	}
}

// Event tests

func TestOnTransitioned(t *testing.T) {
	var transitions []Transition[State, Trigger]
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).Permit(TriggerY, StateC)

	sm.OnTransitioned(func(transition Transition[State, Trigger]) {
		transitions = append(transitions, transition)
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(transitions) != 2 {
		t.Errorf("expected 2 transitions, got %d", len(transitions))
	}
}

func TestOnTransitionCompleted(t *testing.T) {
	completedCount := 0
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.OnTransitionCompleted(func(transition Transition[State, Trigger]) {
		completedCount++
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if completedCount != 1 {
		t.Errorf("expected OnTransitionCompleted to be called once, got %d", completedCount)
	}
}

func TestUnregisterAllCallbacks(t *testing.T) {
	transitionCount := 0
	completedCount := 0

	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).Permit(TriggerY, StateA)

	sm.OnTransitioned(func(transition Transition[State, Trigger]) {
		transitionCount++
	})
	sm.OnTransitionCompleted(func(transition Transition[State, Trigger]) {
		completedCount++
	})

	// Fire once - callbacks should be called
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if transitionCount != 1 {
		t.Errorf("expected transitionCount to be 1, got %d", transitionCount)
	}
	if completedCount != 1 {
		t.Errorf("expected completedCount to be 1, got %d", completedCount)
	}

	// Unregister all callbacks
	sm.UnregisterAllCallbacks()

	// Fire again - callbacks should NOT be called
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if transitionCount != 1 {
		t.Errorf("expected transitionCount to still be 1 after unregister, got %d", transitionCount)
	}
	if completedCount != 1 {
		t.Errorf("expected completedCount to still be 1 after unregister, got %d", completedCount)
	}
}

// External storage tests

func TestExternalStorage(t *testing.T) {
	externalState := StateA

	sm := NewStateMachineWithExternalStorage[State, Trigger](
		func() State { return externalState },
		func(s State) { externalState = s },
	)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if externalState != StateB {
		t.Errorf("expected external state to be StateB, got %v", externalState)
	}
	if sm.State() != StateB {
		t.Errorf("expected sm.State() to be StateB, got %v", sm.State())
	}
}

// Unhandled trigger tests

func TestOnUnhandledTrigger(t *testing.T) {
	var unhandledState State
	var unhandledTrigger Trigger
	var unhandledGuards []string

	sm := NewStateMachine[State, Trigger](StateA)
	sm.OnUnhandledTrigger(func(state State, trigger Trigger, unmetGuards []string) {
		unhandledState = state
		unhandledTrigger = trigger
		unhandledGuards = unmetGuards
	})

	// Fire an unconfigured trigger
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error when OnUnhandledTrigger is set: %v", err)
	}

	if unhandledState != StateA {
		t.Errorf("expected unhandled state to be StateA, got %v", unhandledState)
	}
	if unhandledTrigger != TriggerX {
		t.Errorf("expected unhandled trigger to be TriggerX, got %v", unhandledTrigger)
	}
	if unhandledGuards != nil {
		t.Errorf("expected no unmet guards, got %v", unhandledGuards)
	}
}

// CanFire tests

func TestCanFire(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		Ignore(TriggerY)

	if !sm.CanFire(TriggerX, nil) {
		t.Error("expected CanFire(TriggerX) to be true")
	}
	if !sm.CanFire(TriggerY, nil) {
		t.Error("expected CanFire(TriggerY) to be true (ignored triggers can be fired)")
	}
	if sm.CanFire(TriggerZ, nil) {
		t.Error("expected CanFire(TriggerZ) to be false")
	}
}

func TestCanFire_WithGuard(t *testing.T) {
	guardResult := true
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitIf(TriggerX, StateB, func() bool { return guardResult })

	if !sm.CanFire(TriggerX, nil) {
		t.Error("expected CanFire(TriggerX) to be true when guard passes")
	}

	guardResult = false
	if sm.CanFire(TriggerX, nil) {
		t.Error("expected CanFire(TriggerX) to be false when guard fails")
	}
}

// GetPermittedTriggers tests

func TestGetPermittedTriggers(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		Permit(TriggerY, StateC)

	triggers := sm.GetPermittedTriggers(nil)

	if len(triggers) != 2 {
		t.Errorf("expected 2 permitted triggers, got %d", len(triggers))
	}

	hasTriggerX := false
	hasTriggerY := false
	for _, tr := range triggers {
		if tr == TriggerX {
			hasTriggerX = true
		}
		if tr == TriggerY {
			hasTriggerY = true
		}
	}

	if !hasTriggerX {
		t.Error("expected TriggerX in permitted triggers")
	}
	if !hasTriggerY {
		t.Error("expected TriggerY in permitted triggers")
	}
}

// Firing mode tests

func TestFiringModeQueued(t *testing.T) {
	sm := NewStateMachineWithMode[State, Trigger](StateA, FiringQueued)
	transitions := make([]Transition[State, Trigger], 0)

	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error {
			// Fire another trigger from within an exit action
			go func() {
				sm.Fire(TriggerY, nil)
			}()
			return nil
		})
	sm.Configure(StateB).
		Permit(TriggerY, StateC)

	sm.OnTransitioned(func(transition Transition[State, Trigger]) {
		transitions = append(transitions, transition)
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Wait a bit for the queued event to process
	time.Sleep(100 * time.Millisecond)

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

// Context cancellation test

func TestFireCtx_Cancellation(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := sm.FireCtx(ctx, TriggerX, nil)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Concurrent access test

func TestConcurrentFire(t *testing.T) {
	sm := NewStateMachineWithMode[State, Trigger](StateA, FiringQueued)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		PermitReentry(TriggerY)
	sm.Configure(StateB).
		Permit(TriggerX, StateA).
		PermitReentry(TriggerY)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.Fire(TriggerX, nil)
			sm.Fire(TriggerY, nil)
		}()
	}
	wg.Wait()

	// Just ensure no panics occurred
}

func TestImmediateEntryAProcessedBeforeEnterB(t *testing.T) {
	record := []string{}
	sm := NewStateMachineWithMode[State, Trigger](StateA, FiringImmediate)

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "EnterA"); return nil }).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitA"); return nil })

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
			// Fire this before finishing processing the entry action
			sm.Fire(TriggerY, nil)
			record = append(record, "EnterB")
			return nil
		}).
		Permit(TriggerY, StateA).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitB"); return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected sequence of events: Exit A -> Exit B -> Enter A -> Enter B
	expected := []string{"ExitA", "ExitB", "EnterA", "EnterB"}
	if len(record) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(record), record)
	}
	for i := 0; i < len(expected); i++ {
		if record[i] != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, record[i])
		}
	}
}

func TestQueuedEntryAProcessedAfterEnterB(t *testing.T) {
	record := []string{}
	sm := NewStateMachineWithMode[State, Trigger](StateA, FiringQueued)

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "EnterA"); return nil }).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitA"); return nil })

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
			// Fire this before finishing processing the entry action
			sm.Fire(TriggerY, nil)
			record = append(record, "EnterB")
			return nil
		}).
		Permit(TriggerY, StateA).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitB"); return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected sequence of events: Exit A -> Enter B -> Exit B -> Enter A
	expected := []string{"ExitA", "EnterB", "ExitB", "EnterA"}
	if len(record) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(record), record)
	}
	for i := 0; i < len(expected); i++ {
		if record[i] != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, record[i])
		}
	}
}

func TestImmediateFiringOnEntryEndsUpInCorrectState(t *testing.T) {
	record := []string{}
	sm := NewStateMachineWithMode[State, Trigger](StateA, FiringImmediate)

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "EnterA"); return nil }).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitA"); return nil })

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
			record = append(record, "EnterB")
			// Fire this before finishing processing the entry action
			sm.Fire(TriggerX, nil)
			return nil
		}).
		Permit(TriggerX, StateC).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitB"); return nil })

	sm.Configure(StateC).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "EnterC"); return nil }).
		Permit(TriggerX, StateA).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { record = append(record, "ExitC"); return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected sequence of events: ExitA -> EnterB -> ExitB -> EnterC
	expected := []string{"ExitA", "EnterB", "ExitB", "EnterC"}
	if len(record) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(record), record)
	}
	for i := 0; i < len(expected); i++ {
		if record[i] != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, record[i])
		}
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

// GetInfo test

func TestGetInfo(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { return nil })
	sm.Configure(StateB).
		Permit(TriggerY, StateA)

	info := sm.GetInfo()

	if info == nil {
		t.Fatal("expected non-nil info")
	}
	if len(info.States) != 2 {
		t.Errorf("expected 2 states, got %d", len(info.States))
	}
	if info.InitialState == nil {
		t.Error("expected non-nil initial state")
	}
}

func TestGetInfo_ShouldReturnEntryAction(t *testing.T) {
	// ARRANGE
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { return nil })

	// ACT
	stateMachineInfo := sm.GetInfo()

	// ASSERT
	if len(stateMachineInfo.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(stateMachineInfo.States))
	}
	stateInfo := stateMachineInfo.States[0]
	if len(stateInfo.EntryActions) != 1 {
		t.Fatalf("expected 1 entry action, got %d", len(stateInfo.EntryActions))
	}
}

// String representation test

func TestStateMachine_String(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	str := sm.String()
	if str == "" {
		t.Error("expected non-empty string representation")
	}
}

// Transition tests

func TestTransition_IsReentry(t *testing.T) {
	trans := Transition[State, Trigger]{Source: StateA, Destination: StateA, Trigger: TriggerX}
	if !trans.IsReentry() {
		t.Error("expected IsReentry to be true for same source and destination")
	}

	trans2 := Transition[State, Trigger]{Source: StateA, Destination: StateB, Trigger: TriggerX}
	if trans2.IsReentry() {
		t.Error("expected IsReentry to be false for different source and destination")
	}
}

func TestInitialTransition(t *testing.T) {
	trans := Transition[State, Trigger]{
		Source:      StateA,
		Destination: StateB,
		Trigger:     TriggerX,
		isInitial:   true,
	}
	if !trans.IsInitial() {
		t.Error("expected IsInitial to be true for initial transition")
	}
}

// Initial transition tests (ported from .NET Stateless)

func TestInitialTransition_EntersSubState(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateC)

	sm.Configure(StateC).
		SubstateOf(StateB)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestInitialTransition_EntersSubStateOfSubstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateC)

	sm.Configure(StateC).
		InitialTransition(StateD).
		SubstateOf(StateB)

	sm.Configure(StateD).
		SubstateOf(StateC)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateD {
		t.Errorf("expected StateD, got %v", sm.State())
	}
}

func TestInitialTransition_DoesNotEnterSubStateOfSubstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB)

	sm.Configure(StateC).
		InitialTransition(StateD).
		SubstateOf(StateB)

	sm.Configure(StateD).
		SubstateOf(StateC)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

func TestInitialTransition_DoNotAllowTransitionToSelf(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for self-transition")
		}
	}()

	sm.Configure(StateA).
		InitialTransition(StateA) // Should panic
}

func TestInitialTransition_DoNotAllowTransitionToAnotherSuperstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateA) // Invalid: StateA is not a substate of StateB

	err := sm.Fire(TriggerX, nil)
	if err == nil {
		t.Error("expected error when initial transition target is not a substate")
	}
}

func TestInitialTransition_DoNotAllowMoreThanOneInitialTransition(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateC)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate initial transition")
		}
	}()

	sm.Configure(StateB).
		InitialTransition(StateA) // Should panic - already has initial transition
}

func TestInitialTransition_WithReentry(t *testing.T) {
	// X: Exit A => Enter A => Enter B
	sm := NewStateMachine[State, Trigger](StateA)

	order := 0
	onEntryStateAfired := 0
	onEntryStateBfired := 0
	onExitStateAfired := 0
	onExitStateBfired := 0

	sm.Configure(StateA).
		InitialTransition(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { order++; onEntryStateAfired = order; return nil }).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { order++; onExitStateAfired = order; return nil }).
		PermitReentry(TriggerX)

	sm.Configure(StateB).
		SubstateOf(StateA).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { order++; onEntryStateBfired = order; return nil }).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { order++; onExitStateBfired = order; return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
	if onExitStateBfired != 0 {
		t.Errorf("expected onExitStateBfired to be 0, got %d", onExitStateBfired)
	}
	if onExitStateAfired != 1 {
		t.Errorf("expected onExitStateAfired to be 1, got %d", onExitStateAfired)
	}
	if onEntryStateAfired != 2 {
		t.Errorf("expected onEntryStateAfired to be 2, got %d", onEntryStateAfired)
	}
	if onEntryStateBfired != 3 {
		t.Errorf("expected onEntryStateBfired to be 3, got %d", onEntryStateBfired)
	}
}

func TestInitialTransition_VerifyNotEnterSuperstateWhenDoingInitialTransition(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateC).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { sm.Fire(TriggerY, nil); return nil }).
		Permit(TriggerY, StateD)

	sm.Configure(StateC).
		SubstateOf(StateB).
		Permit(TriggerY, StateD)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateD {
		t.Errorf("expected StateD, got %v", sm.State())
	}
}

func TestInitialTransition_SubStateOfSubstateOnEntryCountAndOrder(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	onEntryCount := ""

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { onEntryCount += "A"; return nil }).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { onEntryCount += "B"; return nil }).
		InitialTransition(StateC)

	sm.Configure(StateC).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { onEntryCount += "C"; return nil }).
		InitialTransition(StateD).
		SubstateOf(StateB)

	sm.Configure(StateD).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { onEntryCount += "D"; return nil }).
		SubstateOf(StateC)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if onEntryCount != "BCD" {
		t.Errorf("expected onEntryCount to be 'BCD', got '%s'", onEntryCount)
	}
}

func TestInitialTransition_TransitionEventsOrdering(t *testing.T) {
	expectedOrdering := []string{"OnExitA", "OnTransitionedStateAStateB", "OnEntryB", "OnTransitionedStateBStateC", "OnEntryC", "OnTransitionCompletedStateAStateC"}
	actualOrdering := []string{}

	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "OnExitA"); return nil })

	sm.Configure(StateB).
		InitialTransition(StateC).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "OnEntryB"); return nil })

	sm.Configure(StateC).
		SubstateOf(StateB).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "OnEntryC"); return nil })

	sm.OnTransitioned(func(t Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned"+t.Source.String()+t.Destination.String())
	})

	sm.OnTransitionCompleted(func(t Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitionCompleted"+t.Source.String()+t.Destination.String())
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}

	if len(expectedOrdering) != len(actualOrdering) {
		t.Fatalf("expected %d events, got %d: %v", len(expectedOrdering), len(actualOrdering), actualOrdering)
	}
	for i := 0; i < len(expectedOrdering); i++ {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected '%s' at index %d, got '%s'", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

// Typed OnEntry tests using type assertion

type AssignArgs struct {
	Assignee string
}

func TestOnEntry_TypedArgument(t *testing.T) {
	var receivedArgs AssignArgs
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	// Configure state with typed entry action using type assertion
	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, trans Transition[State, Trigger]) error {
			if args, ok := trans.Args.(AssignArgs); ok {
				receivedArgs = args
			}
			return nil
		})

	// Fire with typed argument
	err := sm.Fire(TriggerX, AssignArgs{Assignee: "Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedArgs.Assignee != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", receivedArgs.Assignee)
	}
}

func TestOnEntry_CheckTrigger_TypedArgument(t *testing.T) {
	var receivedArgs AssignArgs
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	// Configure state with typed entry action checking specific trigger
	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, trans Transition[State, Trigger]) error {
			if trans.Trigger == TriggerX {
				if args, ok := trans.Args.(AssignArgs); ok {
					receivedArgs = args
				}
			}
			return nil
		})

	// Fire with typed argument
	err := sm.Fire(TriggerX, AssignArgs{Assignee: "Bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedArgs.Assignee != "Bob" {
		t.Errorf("expected 'Bob', got '%s'", receivedArgs.Assignee)
	}
}

func TestOnExit_TypedArgument(t *testing.T) {
	var receivedArgs AssignArgs
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, trans Transition[State, Trigger]) error {
			if args, ok := trans.Args.(AssignArgs); ok {
				receivedArgs = args
			}
			return nil
		})

	sm.Configure(StateB)

	// Fire with typed argument
	err := sm.Fire(TriggerX, AssignArgs{Assignee: "Charlie"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedArgs.Assignee != "Charlie" {
		t.Errorf("expected 'Charlie', got '%s'", receivedArgs.Assignee)
	}
}

// Active states tests (ported from .NET Stateless)

func TestWhenActivate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	expectedOrdering := []string{"ActivatedC", "ActivatedA"}
	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedA"); return nil })

	sm.Configure(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedC"); return nil })

	// should not be called for activation
	sm.OnTransitioned(func(t Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned")
	})
	sm.OnTransitionCompleted(func(t Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitionCompleted")
	})

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expectedOrdering) != len(actualOrdering) {
		t.Fatalf("expected %d events, got %d: %v", len(expectedOrdering), len(actualOrdering), actualOrdering)
	}
	for i := 0; i < len(expectedOrdering); i++ {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

func TestWhenActivateIsIdempotent(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedA"); return nil })

	sm.Configure(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedC"); return nil })

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error on second activate: %v", err)
	}

	if len(actualOrdering) != 2 {
		t.Errorf("expected 2 events, got %d: %v", len(actualOrdering), actualOrdering)
	}
}

func TestWhenDeactivate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	expectedOrdering := []string{"DeactivatedA", "DeactivatedC"}
	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedA"); return nil })

	sm.Configure(StateC).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedC"); return nil })

	// should not be called for deactivation
	sm.OnTransitioned(func(t Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned")
	})
	sm.OnTransitionCompleted(func(t Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitionCompleted")
	})

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Deactivate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expectedOrdering) != len(actualOrdering) {
		t.Fatalf("expected %d events, got %d: %v", len(expectedOrdering), len(actualOrdering), actualOrdering)
	}
	for i := 0; i < len(expectedOrdering); i++ {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

func TestWhenDeactivateIsIdempotent(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedA"); return nil })

	sm.Configure(StateC).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedC"); return nil })

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Deactivate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actualOrdering = []string{} // clear
	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(actualOrdering) != 0 {
		t.Errorf("expected 0 events after re-activate (deactivate should be idempotent), got %d: %v", len(actualOrdering), actualOrdering)
	}
}

func TestWhenTransitioning(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	expectedOrdering := []string{
		"ActivatedA",
		"ExitedA",
		"OnTransitioned",
		"EnteredB",
		"OnTransitionCompleted",
		"ExitedB",
		"OnTransitioned",
		"EnteredA",
		"OnTransitionCompleted",
	}

	actualOrdering := []string{}

	sm.Configure(StateA).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedA"); return nil }).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedA"); return nil }).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "EnteredA"); return nil }).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "ExitedA"); return nil }).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedB"); return nil }).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedB"); return nil }).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "EnteredB"); return nil }).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { actualOrdering = append(actualOrdering, "ExitedB"); return nil }).
		Permit(TriggerY, StateA)

	sm.OnTransitioned(func(tr Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned")
	})
	sm.OnTransitionCompleted(func(tr Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitionCompleted")
	})

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expectedOrdering) != len(actualOrdering) {
		t.Fatalf("expected %d events, got %d.\nExpected: %v\nActual: %v", len(expectedOrdering), len(actualOrdering), expectedOrdering, actualOrdering)
	}
	for i := 0; i < len(expectedOrdering); i++ {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

func TestWhenTransitioningWithinSameSuperstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	expectedOrdering := []string{
		"ActivatedC",
		"ActivatedA",
	}

	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedA"); return nil }).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedA"); return nil }).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		SubstateOf(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedB"); return nil }).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedB"); return nil }).
		Permit(TriggerY, StateA)

	sm.Configure(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedC"); return nil }).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedC"); return nil })

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sm.Fire(TriggerY, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expectedOrdering) != len(actualOrdering) {
		t.Fatalf("expected %d events, got %d.\nExpected: %v\nActual: %v", len(expectedOrdering), len(actualOrdering), expectedOrdering, actualOrdering)
	}
	for i := 0; i < len(expectedOrdering); i++ {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

// Dynamic trigger behaviour tests (ported from .NET Stateless)

func TestPermitDynamic_Selects_Expected_State(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		PermitDynamic(TriggerX, func() State { return StateB })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

type DynamicArgs struct {
	Value int
}

func TestPermitDynamic_With_TriggerParameter_Selects_Expected_State(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		PermitDynamicArgs(TriggerX, func(args any) State {
			if da, ok := args.(DynamicArgs); ok && da.Value == 1 {
				return StateB
			}
			return StateC
		})

	if err := sm.Fire(TriggerX, DynamicArgs{Value: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

func TestPermitDynamic_Permits_Reentry(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	onExitInvoked := false
	onEntryInvoked := false
	onEntryFromTriggerXInvoked := false

	sm.Configure(StateA).
		PermitDynamic(TriggerX, func() State { return StateA }).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
			onEntryInvoked = true
			if tr.Trigger == TriggerX {
				onEntryFromTriggerXInvoked = true
			}
			return nil
		}).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { onExitInvoked = true; return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !onExitInvoked {
		t.Error("expected OnExit to be invoked")
	}
	if !onEntryInvoked {
		t.Error("expected OnEntry to be invoked")
	}
	if !onEntryFromTriggerXInvoked {
		t.Error("expected OnEntry to detect TriggerX")
	}
	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
}

func TestPermitDynamic_Selects_Expected_State_Based_On_DestinationStateSelector_Function(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	value := 'C'
	sm.Configure(StateA).
		PermitDynamic(TriggerX, func() State {
			if value == 'B' {
				return StateB
			}
			return StateC
		})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestPermitDynamicIf_With_TriggerParameter_Permits_Transition_When_GuardCondition_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicArgsIf(TriggerX,
			func(args any) State {
				if da, ok := args.(DynamicArgs); ok && da.Value == 1 {
					return StateC
				}
				return StateB
			},
			func(args any) bool {
				if da, ok := args.(DynamicArgs); ok {
					return da.Value == 1
				}
				return false
			},
		)

	if err := sm.Fire(TriggerX, DynamicArgs{Value: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

type DynamicArgs2 struct {
	I int
	J int
}

func TestPermitDynamicIf_With_2_TriggerParameters_Permits_Transition_When_GuardCondition_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicArgsIf(TriggerX,
			func(args any) State {
				if da, ok := args.(DynamicArgs2); ok && da.I == 1 && da.J == 2 {
					return StateC
				}
				return StateB
			},
			func(args any) bool {
				if da, ok := args.(DynamicArgs2); ok {
					return da.I == 1 && da.J == 2
				}
				return false
			},
		)

	if err := sm.Fire(TriggerX, DynamicArgs2{I: 1, J: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

type DynamicArgs3 struct {
	I int
	J int
	K int
}

func TestPermitDynamicIf_With_3_TriggerParameters_Permits_Transition_When_GuardCondition_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicArgsIf(TriggerX,
			func(args any) State {
				if da, ok := args.(DynamicArgs3); ok && da.I == 1 && da.J == 2 && da.K == 3 {
					return StateC
				}
				return StateB
			},
			func(args any) bool {
				if da, ok := args.(DynamicArgs3); ok {
					return da.I == 1 && da.J == 2 && da.K == 3
				}
				return false
			},
		)

	if err := sm.Fire(TriggerX, DynamicArgs3{I: 1, J: 2, K: 3}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestPermitDynamicIf_With_TriggerParameter_Throws_When_GuardCondition_Not_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicArgsIf(TriggerX,
			func(args any) State {
				if da, ok := args.(DynamicArgs); ok && da.Value > 0 {
					return StateC
				}
				return StateB
			},
			func(args any) bool {
				if da, ok := args.(DynamicArgs); ok {
					return da.Value == 2
				}
				return false
			},
		)

	err := sm.Fire(TriggerX, DynamicArgs{Value: 1})
	if err == nil {
		t.Error("expected error when guard condition not met")
	}
}

func TestPermitDynamicIf_With_2_TriggerParameters_Throws_When_GuardCondition_Not_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicArgsIf(TriggerX,
			func(args any) State {
				if da, ok := args.(DynamicArgs2); ok && da.I > 0 {
					return StateC
				}
				return StateB
			},
			func(args any) bool {
				if da, ok := args.(DynamicArgs2); ok {
					return da.I == 2 && da.J == 3
				}
				return false
			},
		)

	err := sm.Fire(TriggerX, DynamicArgs2{I: 1, J: 2})
	if err == nil {
		t.Error("expected error when guard condition not met")
	}
}

func TestPermitDynamicIf_With_3_TriggerParameters_Throws_When_GuardCondition_Not_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicArgsIf(TriggerX,
			func(args any) State {
				if da, ok := args.(DynamicArgs3); ok && da.I > 0 {
					return StateC
				}
				return StateB
			},
			func(args any) bool {
				if da, ok := args.(DynamicArgs3); ok {
					return da.I == 2 && da.J == 3 && da.K == 4
				}
				return false
			},
		)

	err := sm.Fire(TriggerX, DynamicArgs3{I: 1, J: 2, K: 3})
	if err == nil {
		t.Error("expected error when guard condition not met")
	}
}

func TestPermitDynamicIf_Permits_Reentry_When_GuardCondition_Met(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	onExitInvoked := false
	onEntryInvoked := false
	onEntryFromTriggerXInvoked := false

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX, func() State { return StateA }, func() bool { return true }).
		OnEntry(func(ctx context.Context, tr Transition[State, Trigger]) error {
			onEntryInvoked = true
			if tr.Trigger == TriggerX {
				onEntryFromTriggerXInvoked = true
			}
			return nil
		}).
		OnExit(func(ctx context.Context, tr Transition[State, Trigger]) error { onExitInvoked = true; return nil })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !onExitInvoked {
		t.Error("expected OnExit to be invoked")
	}
	if !onEntryInvoked {
		t.Error("expected OnEntry to be invoked")
	}
	if !onEntryFromTriggerXInvoked {
		t.Error("expected OnEntry to detect TriggerX")
	}
	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
}

// Ignored trigger behaviour tests (ported from .NET Stateless)

func TestIgnore_StateRemainsUnchanged(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Ignore(TriggerX)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
}

func TestIgnoredTriggerBehaviour_ExposesCorrectUnderlyingTrigger(t *testing.T) {
	ignored := NewIgnoredTriggerBehaviour[State, Trigger](TriggerX, EmptyTransitionGuard)

	if ignored.GetTrigger() != TriggerX {
		t.Errorf("expected TriggerX, got %v", ignored.GetTrigger())
	}
}

func TestIgnoredTriggerBehaviour_WhenGuardConditionFalse_IsGuardConditionMetIsFalse(t *testing.T) {
	guardFalse := func() bool { return false }
	ignored := NewIgnoredTriggerBehaviour[State, Trigger](TriggerX, NewTransitionGuard(guardFalse, ""))

	if ignored.GuardConditionsMet(nil) {
		t.Error("expected GuardConditionsMet to be false")
	}
}

func TestIgnoredTriggerBehaviour_WhenGuardConditionTrue_IsGuardConditionMetIsTrue(t *testing.T) {
	guardTrue := func() bool { return true }
	ignored := NewIgnoredTriggerBehaviour[State, Trigger](TriggerX, NewTransitionGuard(guardTrue, ""))

	if !ignored.GuardConditionsMet(nil) {
		t.Error("expected GuardConditionsMet to be true")
	}
}

func TestIgnoredTriggerMustBeIgnoredSync(t *testing.T) {
	// In a substate hierarchy, ignored trigger in substate should be properly ignored
	// and not cause the superstate's transition to execute
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		Permit(TriggerX, StateC)

	sm.Configure(StateB).
		SubstateOf(StateA).
		Ignore(TriggerX)

	// This should not panic and should stay in StateB
	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB (trigger should be ignored), got %v", sm.State())
	}
}

func TestIgnoreIfTrueTriggerMustBeIgnored(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		Permit(TriggerX, StateC)

	sm.Configure(StateB).
		SubstateOf(StateA).
		IgnoreIf(TriggerX, func() bool { return true })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

func TestIgnoreIfFalseTriggerMustNotBeIgnored(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		Permit(TriggerX, StateC)

	sm.Configure(StateB).
		SubstateOf(StateA).
		IgnoreIf(TriggerX, func() bool { return false })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestInternalTransitionIf_ShouldExecuteOnlyFirstMatchingAction(t *testing.T) {
	sm := NewStateMachine[int, int](1)
	executed := false

	sm.Configure(1).
		InternalTransitionIf(1, func() bool { return true }, func(ctx context.Context, tr Transition[int, int]) error {
			executed = true
			return nil
		}).
		InternalTransitionIf(1, func() bool { return false }, func(ctx context.Context, tr Transition[int, int]) error {
			t.Error("second action should not be executed")
			return nil
		})

	if err := sm.Fire(1, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !executed {
		t.Error("first action should have been executed")
	}
}

func TestSubstateTransition_OverridesSuperstate(t *testing.T) {
	sm := NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB)

	// Overrides the superstate transition
	sm.Configure(StateB).
		SubstateOf(StateA).
		Permit(TriggerX, StateC)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestSubstateTransition_GuardBlocked_UsesSuperstateTransition(t *testing.T) {
	guardConditionValue := false
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		PermitIf(TriggerX, StateD, func() bool { return true })

	sm.Configure(StateB).
		SubstateOf(StateA).
		PermitIf(TriggerX, StateC, func() bool { return guardConditionValue })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateD {
		t.Errorf("expected StateD (superstate transition), got %v", sm.State())
	}
}

func TestSubstateTransition_GuardOpen_UsesSubstateTransition(t *testing.T) {
	guardConditionValue := true
	sm := NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		PermitIf(TriggerX, StateD, func() bool { return true })

	sm.Configure(StateB).
		SubstateOf(StateA).
		PermitIf(TriggerX, StateC, func() bool { return guardConditionValue })

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC (substate transition), got %v", sm.State())
	}
}

func TestMultiLayerSubstates_GuardFallthrough(t *testing.T) {
	testCases := []struct {
		name                       string
		parentGuard                bool
		childGuard                 bool
		grandchildGuard            bool
		expectedState              string
	}{
		{"grandchild open only", false, false, true, "GrandchildStateTarget"},
		{"child open only", false, true, false, "ChildStateTarget"},
		{"child and grandchild open", false, true, true, "GrandchildStateTarget"},
		{"parent open only", true, false, false, "ParentStateTarget"},
		{"parent and grandchild open", true, false, true, "GrandchildStateTarget"},
		{"parent and child open", true, true, false, "ChildStateTarget"},
		{"all open", true, true, true, "GrandchildStateTarget"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewStateMachine[string, Trigger]("GrandchildState")

			sm.Configure("ParentState").
				PermitIf(TriggerX, "ParentStateTarget", func() bool { return tc.parentGuard })

			sm.Configure("ChildState").
				SubstateOf("ParentState").
				PermitIf(TriggerX, "ChildStateTarget", func() bool { return tc.childGuard })

			sm.Configure("GrandchildState").
				SubstateOf("ChildState").
				PermitIf(TriggerX, "GrandchildStateTarget", func() bool { return tc.grandchildGuard })

			if err := sm.Fire(TriggerX, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if sm.State() != tc.expectedState {
				t.Errorf("expected %s, got %s", tc.expectedState, sm.State())
			}
		})
	}
}

// TestChildToParentTransition_OnEntryNotFired documents the behavior when
// transitioning from a child state to its parent state: OnEntry is NOT fired
// for the parent because hierarchically you never "left" the parent state.
// This matches .NET Stateless and qmuntal/stateless behavior.
// See: https://github.com/qmuntal/stateless/issues/98
//
// If you need OnEntry to fire, use PermitReentry instead of Permit.
func TestChildToParentTransition_OnEntryNotFired(t *testing.T) {
	type Issue98State string
	type Issue98Trigger string

	const (
		working   Issue98State   = "Working"
		substateA Issue98State   = "SubstateA"
		substateB Issue98State   = "SubstateB"
		goToA     Issue98Trigger = "GoToA"
		goToB     Issue98Trigger = "GoToB"
		exitA     Issue98Trigger = "ExitA"
	)

	sm := NewStateMachine[Issue98State, Issue98Trigger](working)

	var workingEntryCount int
	var substateAExitCount int

	sm.Configure(working).
		OnEntry(func(ctx context.Context, tr Transition[Issue98State, Issue98Trigger]) error {
			workingEntryCount++
			return nil
		}).
		Permit(goToA, substateA).
		Permit(goToB, substateB)

	sm.Configure(substateA).
		SubstateOf(working).
		OnExit(func(ctx context.Context, tr Transition[Issue98State, Issue98Trigger]) error {
			substateAExitCount++
			return nil
		}).
		Permit(exitA, working)

	sm.Configure(substateB).
		SubstateOf(working)

	// Go to substateA
	if err := sm.Fire(goToA, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Transition from substateA back to working (parent state)
	if err := sm.Fire(exitA, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Current behavior: OnEntry for parent is NOT called when coming from child
	// This matches qmuntal/stateless and dotnet/stateless behavior
	if workingEntryCount != 0 {
		t.Errorf("expected workingEntryCount to be 0 (parent OnEntry not called), got %d", workingEntryCount)
	}

	// But OnExit for child IS called
	if substateAExitCount != 1 {
		t.Errorf("expected substateAExitCount to be 1, got %d", substateAExitCount)
	}

	// State should be working
	if sm.State() != working {
		t.Errorf("expected working, got %v", sm.State())
	}
}
