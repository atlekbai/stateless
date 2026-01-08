package stateless_test

import (
	"context"
	"errors"
	"testing"

	"github.com/atlekbai/stateless"
)

// Test state and trigger types.
type (
	State   int
	Trigger int
)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	if sm.State() != StateA {
		t.Errorf("expected initial state to be StateA, got %v", sm.State())
	}
}

func TestSimpleTransition(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected state to be StateB, got %v", sm.State())
	}
}

func TestMultipleTransitions(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	// TriggerY is not configured for StateA
	err := sm.Fire(TriggerY, nil)
	if err == nil {
		t.Error("expected error for invalid transition")
	}

	var invalidTransitionErr *stateless.InvalidTransitionError
	if !errors.As(err, &invalidTransitionErr) {
		t.Errorf("expected InvalidTransitionError, got %T", err)
	}
}

// GetPermittedTriggers tests

func TestGetPermittedTriggers(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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

// GetInfo test

func TestGetInfo(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil })
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil })

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	str := sm.String()
	if str == "" {
		t.Error("expected non-empty string representation")
	}
}

// Unhandled trigger tests

func TestOnUnhandledTrigger(t *testing.T) {
	var unhandledState State
	var unhandledTrigger Trigger
	var unhandledGuards []string

	sm := stateless.NewStateMachine[State, Trigger](StateA)
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

// External storage tests

func TestExternalStorage(t *testing.T) {
	externalState := StateA

	sm := stateless.NewStateMachineWithExternalStorage[State, Trigger](
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

// Activation/Deactivation tests

func TestActivateDeactivate(t *testing.T) {
	activateCount := 0
	deactivateCount := 0

	sm := stateless.NewStateMachine[State, Trigger](StateA)
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

// Active states tests (ported from .NET Stateless)

func TestWhenActivate(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	expectedOrdering := []string{"ActivatedC", "ActivatedA"}
	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedA"); return nil })

	sm.Configure(StateC).
		OnActivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "ActivatedC"); return nil })

	// should not be called for activation
	sm.OnTransitioned(func(t stateless.Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned")
	})
	sm.OnTransitionCompleted(func(t stateless.Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitionCompleted")
	})

	if err := sm.Activate(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(expectedOrdering) != len(actualOrdering) {
		t.Fatalf("expected %d events, got %d: %v", len(expectedOrdering), len(actualOrdering), actualOrdering)
	}
	for i := range expectedOrdering {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

func TestWhenActivateIsIdempotent(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	expectedOrdering := []string{"DeactivatedA", "DeactivatedC"}
	actualOrdering := []string{}

	sm.Configure(StateA).
		SubstateOf(StateC).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedA"); return nil })

	sm.Configure(StateC).
		OnDeactivate(func(ctx context.Context) error { actualOrdering = append(actualOrdering, "DeactivatedC"); return nil })

	// should not be called for deactivation
	sm.OnTransitioned(func(t stateless.Transition[State, Trigger]) {
		actualOrdering = append(actualOrdering, "OnTransitioned")
	})
	sm.OnTransitionCompleted(func(t stateless.Transition[State, Trigger]) {
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
	for i := range expectedOrdering {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}

func TestWhenDeactivateIsIdempotent(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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

// CanFire tests

func TestCanFire(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitIf(TriggerX, StateB, func() bool { return guardResult })

	if !sm.CanFire(TriggerX, nil) {
		t.Error("expected CanFire(TriggerX) to be true when guard passes")
	}

	guardResult = false
	if sm.CanFire(TriggerX, nil) {
		t.Error("expected CanFire(TriggerX) to be false when guard fails")
	}
}

// Guard tests

func TestPermitIf_GuardPasses(t *testing.T) {
	guardCalled := false
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
