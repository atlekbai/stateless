package stateless_test

import (
	"context"
	"testing"

	"github.com/atlekbai/stateless"
)

func TestSubstateOf(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateC)
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

func TestSubstateTransition_OverridesSuperstate(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateB)

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
	sm := stateless.NewStateMachine[State, Trigger](StateB)

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
		name            string
		parentGuard     bool
		childGuard      bool
		grandchildGuard bool
		expectedState   string
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
			sm := stateless.NewStateMachine[string, Trigger]("GrandchildState")

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

	sm := stateless.NewStateMachine[Issue98State, Issue98Trigger](working)

	var workingEntryCount int
	var substateAExitCount int

	sm.Configure(working).
		OnEntry(func(ctx context.Context, tr stateless.Transition[Issue98State, Issue98Trigger]) error {
			workingEntryCount++
			return nil
		}).
		Permit(goToA, substateA).
		Permit(goToB, substateB)

	sm.Configure(substateA).
		SubstateOf(working).
		OnExit(func(ctx context.Context, tr stateless.Transition[Issue98State, Issue98Trigger]) error {
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

func TestWhenTransitioningWithinSameSuperstate(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	for i := range expectedOrdering {
		if expectedOrdering[i] != actualOrdering[i] {
			t.Errorf("expected %s at index %d, got %s", expectedOrdering[i], i, actualOrdering[i])
		}
	}
}
