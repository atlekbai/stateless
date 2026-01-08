package stateless_test

import (
	"context"
	"testing"

	"github.com/atlekbai/stateless"
)

func TestOnEntry(t *testing.T) {
	entryCount := 0
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)
	sm.Configure(StateB).OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
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
	var receivedTransition stateless.Transition[State, Trigger]
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).OnEntry(func(ctx context.Context, transition stateless.Transition[State, Trigger]) error {
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

	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).Permit(TriggerX, StateB)

	sm.Configure(StateB).
		Permit(TriggerY, StateC).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
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

// Typed OnEntry tests using type assertion

type AssignArgs struct {
	Assignee string
}

func TestOnEntry_TypedArgument(t *testing.T) {
	var receivedArgs AssignArgs
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	// Configure state with typed entry action using type assertion
	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, trans stateless.Transition[State, Trigger]) error {
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).Permit(TriggerX, StateB)

	// Configure state with typed entry action checking specific trigger
	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, trans stateless.Transition[State, Trigger]) error {
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, trans stateless.Transition[State, Trigger]) error {
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
