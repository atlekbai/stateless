package stateless_test

import (
	"context"
	"testing"

	"github.com/atlekbai/stateless"
)

func TestOnTransitioned(t *testing.T) {
	var transitions []stateless.Transition[State, Trigger]
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).Permit(TriggerY, StateC)

	sm.OnTransitioned(func(transition stateless.Transition[State, Trigger]) {
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.OnTransitionCompleted(func(transition stateless.Transition[State, Trigger]) {
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

	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).Permit(TriggerY, StateA)

	sm.OnTransitioned(func(transition stateless.Transition[State, Trigger]) {
		transitionCount++
	})
	sm.OnTransitionCompleted(func(transition stateless.Transition[State, Trigger]) {
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

func TestWhenTransitioning(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "EnteredA")
			return nil
		}).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "ExitedA")
			return nil
		}).
		Permit(TriggerX, StateB)

	sm.Configure(StateB).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedB"); return nil }).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedB"); return nil }).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "EnteredB")
			return nil
		}).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actualOrdering = append(actualOrdering, "ExitedB")
			return nil
		}).
		Permit(TriggerY, StateA)

	sm.OnTransitioned(func(tr stateless.Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned")
	})
	sm.OnTransitionCompleted(func(tr stateless.Transition[State, Trigger]) {
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
		t.Fatalf(
			"expected %d events, got %d.\nExpected: %v\nActual: %v",
			len(expectedOrdering),
			len(actualOrdering),
			expectedOrdering,
			actualOrdering,
		)
	}
	for i := range expectedOrdering {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}
