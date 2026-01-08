package stateless_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/atlekbai/stateless"
)

// Firing mode tests

func TestFiringModeQueued(t *testing.T) {
	sm := stateless.NewStateMachineWithMode[State, Trigger](StateA, stateless.FiringQueued)
	transitions := make([]stateless.Transition[State, Trigger], 0)

	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			// Fire another trigger from within an exit action
			go func() {
				sm.Fire(TriggerY, nil)
			}()
			return nil
		})
	sm.Configure(StateB).
		Permit(TriggerY, StateC)

	sm.OnTransitioned(func(transition stateless.Transition[State, Trigger]) {
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachineWithMode[State, Trigger](StateA, stateless.FiringQueued)
	sm.Configure(StateA).
		Permit(TriggerX, StateB).
		PermitReentry(TriggerY)
	sm.Configure(StateB).
		Permit(TriggerX, StateA).
		PermitReentry(TriggerY)

	var wg sync.WaitGroup
	for range 100 {
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
	sm := stateless.NewStateMachineWithMode[State, Trigger](StateA, stateless.FiringImmediate)

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "EnterA")
			return nil
		}).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitA")
			return nil
		})

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			// Fire this before finishing processing the entry action
			sm.Fire(TriggerY, nil)
			record = append(record, "EnterB")
			return nil
		}).
		Permit(TriggerY, StateA).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitB")
			return nil
		})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected sequence of events: Exit A -> Exit B -> Enter A -> Enter B
	expected := []string{"ExitA", "ExitB", "EnterA", "EnterB"}
	if len(record) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(record), record)
	}
	for i := range expected {
		if record[i] != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, record[i])
		}
	}
}

func TestQueuedEntryAProcessedAfterEnterB(t *testing.T) {
	record := []string{}
	sm := stateless.NewStateMachineWithMode[State, Trigger](StateA, stateless.FiringQueued)

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "EnterA")
			return nil
		}).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitA")
			return nil
		})

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			// Fire this before finishing processing the entry action
			sm.Fire(TriggerY, nil)
			record = append(record, "EnterB")
			return nil
		}).
		Permit(TriggerY, StateA).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitB")
			return nil
		})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected sequence of events: Exit A -> Enter B -> Exit B -> Enter A
	expected := []string{"ExitA", "EnterB", "ExitB", "EnterA"}
	if len(record) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(record), record)
	}
	for i := range expected {
		if record[i] != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, record[i])
		}
	}
}

func TestImmediateFiringOnEntryEndsUpInCorrectState(t *testing.T) {
	record := []string{}
	sm := stateless.NewStateMachineWithMode[State, Trigger](StateA, stateless.FiringImmediate)

	sm.Configure(StateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "EnterA")
			return nil
		}).
		Permit(TriggerX, StateB).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitA")
			return nil
		})

	sm.Configure(StateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "EnterB")
			// Fire this before finishing processing the entry action
			sm.Fire(TriggerX, nil)
			return nil
		}).
		Permit(TriggerX, StateC).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitB")
			return nil
		})

	sm.Configure(StateC).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "EnterC")
			return nil
		}).
		Permit(TriggerX, StateA).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			record = append(record, "ExitC")
			return nil
		})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected sequence of events: ExitA -> EnterB -> ExitB -> EnterC
	expected := []string{"ExitA", "EnterB", "ExitB", "EnterC"}
	if len(record) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(record), record)
	}
	for i := range expected {
		if record[i] != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, record[i])
		}
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}
