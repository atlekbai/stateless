package stateless_test

import (
	"context"
	"errors"
	"testing"

	"github.com/atlekbai/stateless"
)

// Reentry tests

func TestPermitReentry(t *testing.T) {
	entryCount := 0
	exitCount := 0

	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		PermitReentry(TriggerX).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { entryCount++; return nil }).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error { exitCount++; return nil })

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

// Dynamic transition tests

func TestPermitDynamic(t *testing.T) {
	destState := StateB
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitDynamic(TriggerX, func(_ context.Context, _ any) (State, error) {
		return destState, nil
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}

	// Reset and try with different destination
	destState = StateC
	sm = stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitDynamic(TriggerX, func(_ context.Context, _ any) (State, error) {
		return destState, nil
	})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestPermitDynamicWithArgs(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).PermitDynamic(TriggerX, func(_ context.Context, args any) (State, error) {
		if state, ok := args.(State); ok {
			return state, nil
		}
		return StateB, nil
	})

	if err := sm.Fire(TriggerX, StateC); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

// Dynamic trigger behaviour tests (ported from .NET Stateless)

func TestPermitDynamic_Selects_Expected_State(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		PermitDynamic(TriggerX, func(_ context.Context, _ any) (State, error) { return StateB, nil })

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	sm.Configure(StateA).
		PermitDynamic(TriggerX, func(_ context.Context, args any) (State, error) {
			if da, ok := args.(DynamicArgs); ok && da.Value == 1 {
				return StateB, nil
			}
			return StateC, nil
		})

	if err := sm.Fire(TriggerX, DynamicArgs{Value: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateB {
		t.Errorf("expected StateB, got %v", sm.State())
	}
}

func TestPermitDynamic_Permits_Reentry(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	onExitInvoked := false
	onEntryInvoked := false
	onEntryFromTriggerXInvoked := false

	sm.Configure(StateA).
		PermitDynamic(TriggerX, func(_ context.Context, _ any) (State, error) { return StateA, nil }).
		OnEntry(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			onEntryInvoked = true
			if tr.Trigger == TriggerX {
				onEntryFromTriggerXInvoked = true
			}
			return nil
		}).
		OnExit(func(ctx context.Context, tr stateless.Transition[State, Trigger]) error {
			onExitInvoked = true
			return nil
		})

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	value := 'C'
	sm.Configure(StateA).
		PermitDynamic(TriggerX, func(_ context.Context, _ any) (State, error) {
			if value == 'B' {
				return StateB, nil
			}
			return StateC, nil
		})

	if err := sm.Fire(TriggerX, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.State() != StateC {
		t.Errorf("expected StateC, got %v", sm.State())
	}
}

func TestPermitDynamicIf_With_TriggerParameter_Permits_Transition_When_GuardCondition_Met(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX,
			func(_ context.Context, args any) (State, error) {
				if da, ok := args.(DynamicArgs); ok && da.Value == 1 {
					return StateC, nil
				}
				return StateB, nil
			},
			func(ctx context.Context, args any) error {
				if da, ok := args.(DynamicArgs); ok && da.Value == 1 {
					return nil
				}
				return errors.New("guard failed")
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX,
			func(_ context.Context, args any) (State, error) {
				if da, ok := args.(DynamicArgs2); ok && da.I == 1 && da.J == 2 {
					return StateC, nil
				}
				return StateB, nil
			},
			func(ctx context.Context, args any) error {
				if da, ok := args.(DynamicArgs2); ok && da.I == 1 && da.J == 2 {
					return nil
				}
				return errors.New("guard failed")
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX,
			func(_ context.Context, args any) (State, error) {
				if da, ok := args.(DynamicArgs3); ok && da.I == 1 && da.J == 2 && da.K == 3 {
					return StateC, nil
				}
				return StateB, nil
			},
			func(ctx context.Context, args any) error {
				if da, ok := args.(DynamicArgs3); ok && da.I == 1 && da.J == 2 && da.K == 3 {
					return nil
				}
				return errors.New("guard failed")
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX,
			func(_ context.Context, args any) (State, error) {
				if da, ok := args.(DynamicArgs); ok && da.Value > 0 {
					return StateC, nil
				}
				return StateB, nil
			},
			func(ctx context.Context, args any) error {
				if da, ok := args.(DynamicArgs); ok && da.Value == 2 {
					return nil
				}
				return errors.New("guard failed: value must be 2")
			},
		)

	err := sm.Fire(TriggerX, DynamicArgs{Value: 1})
	if err == nil {
		t.Error("expected error when guard condition not met")
	}
}

func TestPermitDynamicIf_With_2_TriggerParameters_Throws_When_GuardCondition_Not_Met(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX,
			func(_ context.Context, args any) (State, error) {
				if da, ok := args.(DynamicArgs2); ok && da.I > 0 {
					return StateC, nil
				}
				return StateB, nil
			},
			func(ctx context.Context, args any) error {
				if da, ok := args.(DynamicArgs2); ok && da.I == 2 && da.J == 3 {
					return nil
				}
				return errors.New("guard failed")
			},
		)

	err := sm.Fire(TriggerX, DynamicArgs2{I: 1, J: 2})
	if err == nil {
		t.Error("expected error when guard condition not met")
	}
}

func TestPermitDynamicIf_With_3_TriggerParameters_Throws_When_GuardCondition_Not_Met(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX,
			func(_ context.Context, args any) (State, error) {
				if da, ok := args.(DynamicArgs3); ok && da.I > 0 {
					return StateC, nil
				}
				return StateB, nil
			},
			func(ctx context.Context, args any) error {
				if da, ok := args.(DynamicArgs3); ok && da.I == 2 && da.J == 3 && da.K == 4 {
					return nil
				}
				return errors.New("guard failed")
			},
		)

	err := sm.Fire(TriggerX, DynamicArgs3{I: 1, J: 2, K: 3})
	if err == nil {
		t.Error("expected error when guard condition not met")
	}
}

func TestPermitDynamicIf_Permits_Reentry_When_GuardCondition_Met(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	onExitInvoked := false
	onEntryInvoked := false
	onEntryFromTriggerXInvoked := false

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX, func(_ context.Context, _ any) (State, error) { return StateA, nil }, func(_ context.Context, _ any) error { return nil }).
		OnEntry(func(_ context.Context, tr stateless.Transition[State, Trigger]) error {
			onEntryInvoked = true
			if tr.Trigger == TriggerX {
				onEntryFromTriggerXInvoked = true
			}
			return nil
		}).
		OnExit(func(_ context.Context, tr stateless.Transition[State, Trigger]) error {
			onExitInvoked = true
			return nil
		})

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
