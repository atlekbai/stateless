package stateless_test

import (
	"context"
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
	sm = stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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

// Dynamic trigger behaviour tests (ported from .NET Stateless)

func TestPermitDynamic_Selects_Expected_State(t *testing.T) {
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	onExitInvoked := false
	onEntryInvoked := false
	onEntryFromTriggerXInvoked := false

	sm.Configure(StateA).
		PermitDynamic(TriggerX, func() State { return StateA }).
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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)

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
	sm := stateless.NewStateMachine[State, Trigger](StateA)
	onExitInvoked := false
	onEntryInvoked := false
	onEntryFromTriggerXInvoked := false

	sm.Configure(StateA).
		PermitDynamicIf(TriggerX, func() State { return StateA }, func() bool { return true }).
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
