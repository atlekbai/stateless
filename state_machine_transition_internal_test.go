package stateless_test

import (
	"context"
	"testing"

	"github.com/atlekbai/stateless"
)

// Internal transition tests

func TestInternalTransition(t *testing.T) {
	actionCount := 0
	entryCount := 0
	exitCount := 0

	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			actionCount++
			return nil
		}).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { entryCount++; return nil }).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { exitCount++; return nil })

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil })

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil }).
		Permit(TriggerY, StateB)

	sm.Configure(StateB).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil }).
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
	sm := stateless.NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil }).
		InternalTransition(TriggerY, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil })

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
	sm := stateless.NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA)

	sm.Configure(StateB).
		SubstateOf(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil }).
		InternalTransition(TriggerY, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil })

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		InternalTransitionIf(TriggerX, func(_ context.Context, _ any) error {
			if !isPermitted {
				return stateless.Reject("not permitted")
			}
			return nil
		}, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { return nil })

	triggers := sm.GetPermittedTriggers(context.Background(), nil)
	if len(triggers) != 1 {
		t.Errorf("expected 1 permitted trigger, got %d", len(triggers))
	}

	isPermitted = false
	triggers = sm.GetPermittedTriggers(context.Background(), nil)
	if len(triggers) != 0 {
		t.Errorf("expected 0 permitted triggers, got %d", len(triggers))
	}
}

func TestInternalTransition_HandledOnlyOnceInSuper(t *testing.T) {
	handledIn := StateC

	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			handledIn = StateA
			return nil
		})

	sm.Configure(StateB).
		SubstateOf(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			handledIn = StateB
			return nil
		})

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

	sm := stateless.NewStateMachine[State, Trigger](StateB)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			handledIn = StateA
			return nil
		})

	sm.Configure(StateB).
		SubstateOf(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			handledIn = StateB
			return nil
		})

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

	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		InternalTransition(TriggerX, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { handled++; return nil }).
		InternalTransition(TriggerY, func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { handled++; return nil })

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

	sm := stateless.NewStateMachine[State, Trigger](StateA)
	var receivedValue int

	sm.Configure(StateA).InternalTransition(TriggerX,
		func(ctx context.Context, trans stateless.Transition[State, Trigger]) error {
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

func TestInternalTransitionIf_ShouldExecuteOnlyFirstMatchingAction(t *testing.T) {
	sm := stateless.NewStateMachine[int, int](1)
	executed := false

	sm.Configure(1).
		InternalTransitionIf(1, func(_ context.Context, _ any) error { return nil }, func(ctx context.Context, tr stateless.Transition[int, int]) error {
			executed = true
			return nil
		}).
		InternalTransitionIf(1, func(_ context.Context, _ any) error { return stateless.Reject("guard failed") }, func(ctx context.Context, tr stateless.Transition[int, int]) error {
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
