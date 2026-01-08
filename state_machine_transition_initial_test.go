package stateless_test

import (
	"context"
	"testing"

	"github.com/atlekbai/stateless"
)

func TestInitialTransition(t *testing.T) {
	trans := stateless.NewInitialTransition[State, Trigger](StateA, StateB, TriggerX, nil)
	if !trans.IsInitial() {
		t.Error("expected IsInitial to be true for initial transition")
	}
}

// Initial transition tests (ported from .NET Stateless)

func TestInitialTransition_EntersSubState(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for self-transition")
		}
	}()

	sm.Configure(StateA).
		InitialTransition(StateA) // Should panic
}

func TestInitialTransition_DoNotAllowTransitionToAnotherSuperstate(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateA) // Invalid: StateA is not a substate of StateB

	err := sm.Fire(TriggerX, nil)
	if err == nil {
		t.Error("expected error when initial transition target is not a substate")
	}
}

func TestInitialTransition_DoNotAllowMoreThanOneInitialTransition(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	order := 0
	onEntryStateAfired := 0
	onEntryStateBfired := 0
	onExitStateAfired := 0
	onExitStateBfired := 0

	sm.Configure(StateA).
		InitialTransition(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			order++
			onEntryStateAfired = order
			return nil
		}).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			order++
			onExitStateAfired = order
			return nil
		}).
		PermitReentry(TriggerX)

	sm.Configure(StateB).
		SubstateOf(StateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			order++
			onEntryStateBfired = order
			return nil
		}).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			order++
			onExitStateBfired = order
			return nil
		})

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		InitialTransition(StateC).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			sm.Fire(TriggerY, nil)
			return nil
		}).
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	onEntryCount := ""

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			onEntryCount += "A"
			return nil
		}).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			onEntryCount += "B"
			return nil
		}).
		InitialTransition(StateC)

	sm.Configure(StateC).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			onEntryCount += "C"
			return nil
		}).
		InitialTransition(StateD).
		SubstateOf(StateB)

	sm.Configure(StateD).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			onEntryCount += "D"
			return nil
		}).
		SubstateOf(StateC)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if onEntryCount != "BCD" {
		t.Errorf("expected onEntryCount to be 'BCD', got '%s'", onEntryCount)
	}
}

func TestInitialTransition_TransitionEventsOrdering(t *testing.T) {
	expectedOrdering := []string{
		"OnExitA",
		"OnTransitionedStateAStateB",
		"OnEntryB",
		"OnTransitionedStateBStateC",
		"OnEntryC",
		"OnTransitionCompletedStateAStateC",
	}
	actualOrdering := []string{}

	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "OnExitA")
			return nil
		})

	sm.Configure(StateB).
		InitialTransition(StateC).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "OnEntryB")
			return nil
		})

	sm.Configure(StateC).
		SubstateOf(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "OnEntryC")
			return nil
		})

	sm.OnTransitioned(func(t stateless.Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned"+t.Source.String()+t.Destination.String())
	})

	sm.OnTransitionCompleted(func(t stateless.Transition[State, Trigger]) {
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
	for i := range expectedOrdering {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected '%s' at index %d, got '%s'", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}
