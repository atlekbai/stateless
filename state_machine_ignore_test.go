package stateless_test

import (
	"testing"

	"github.com/atlekbai/stateless"
)

func TestIgnore(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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

func TestIgnore_StateRemainsUnchanged(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Ignore(TriggerX)

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateA {
		t.Errorf("expected StateA, got %v", sm.State())
	}
}

func TestIgnoredTriggerBehaviour_ExposesCorrectUnderlyingTrigger(t *testing.T) {
	ignored := stateless.NewIgnoredTriggerBehaviour[State, Trigger](TriggerX, stateless.EmptyTransitionGuard)

	if ignored.GetTrigger() != TriggerX {
		t.Errorf("expected TriggerX, got %v", ignored.GetTrigger())
	}
}

func TestIgnoredTriggerBehaviour_WhenGuardConditionFalse_IsGuardConditionMetIsFalse(t *testing.T) {
	guardFalse := func() bool { return false }
	ignored := stateless.NewIgnoredTriggerBehaviour[State, Trigger](TriggerX, stateless.NewTransitionGuard(guardFalse, ""))

	if ignored.GuardConditionsMet(nil) {
		t.Error("expected GuardConditionsMet to be false")
	}
}

func TestIgnoredTriggerBehaviour_WhenGuardConditionTrue_IsGuardConditionMetIsTrue(t *testing.T) {
	guardTrue := func() bool { return true }
	ignored := stateless.NewIgnoredTriggerBehaviour[State, Trigger](TriggerX, stateless.NewTransitionGuard(guardTrue, ""))

	if !ignored.GuardConditionsMet(nil) {
		t.Error("expected GuardConditionsMet to be true")
	}
}

func TestIgnoredTriggerMustBeIgnoredSync(t *testing.T) {
	// In a substate hierarchy, ignored trigger in substate should be properly ignored
	// and not cause the superstate's transition to execute
	sm := stateless.NewStateMachine[State, Trigger](StateB)

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
	sm := stateless.NewStateMachine[State, Trigger](StateB)

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
	sm := stateless.NewStateMachine[State, Trigger](StateB)

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
