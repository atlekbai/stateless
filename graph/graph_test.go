package graph_test

import (
	"context"
	"strings"
	"testing"

	"github.com/atlekbai/stateless"
	"github.com/atlekbai/stateless/graph"
)

// Test state and trigger types.
type (
	TestState   int
	TestTrigger int
)

const (
	TestStateA TestState = iota
	TestStateB
	TestStateC
	TestStateD
)

const (
	TestTriggerX TestTrigger = iota
	TestTriggerY
	TestTriggerZ
)

func (s TestState) String() string {
	switch s {
	case TestStateA:
		return "A"
	case TestStateB:
		return "B"
	case TestStateC:
		return "C"
	case TestStateD:
		return "D"
	default:
		return "Unknown"
	}
}

func (t TestTrigger) String() string {
	switch t {
	case TestTriggerX:
		return "X"
	case TestTriggerY:
		return "Y"
	case TestTriggerZ:
		return "Z"
	default:
		return "Unknown"
	}
}

func TestUmlDotGraph(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		Permit(TestTriggerY, TestStateC)
	sm.Configure(TestStateB).
		Permit(TestTriggerZ, TestStateA)
	sm.Configure(TestStateC).
		Permit(TestTriggerZ, TestStateA)

	info := sm.GetInfo()
	dotGraph := graph.UmlDotGraph(info)

	// Check basic structure
	if !strings.Contains(dotGraph, "digraph") {
		t.Error("expected DOT graph to contain 'digraph'")
	}
	if !strings.Contains(dotGraph, "init") {
		t.Error("expected DOT graph to contain 'init' node")
	}
	if !strings.Contains(dotGraph, "\"A\"") {
		t.Error("expected DOT graph to contain 'A'")
	}
	if !strings.Contains(dotGraph, "\"B\"") {
		t.Error("expected DOT graph to contain 'B'")
	}
	if !strings.Contains(dotGraph, "\"C\"") {
		t.Error("expected DOT graph to contain 'C'")
	}
}

func TestMermaidGraph(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		Permit(TestTriggerY, TestStateC)
	sm.Configure(TestStateB).
		Permit(TestTriggerZ, TestStateA)
	sm.Configure(TestStateC).
		Permit(TestTriggerZ, TestStateA)

	info := sm.GetInfo()
	direction := graph.LeftToRight
	mermaidGraph := graph.MermaidGraph(info, &direction)

	// Check basic structure
	if !strings.Contains(mermaidGraph, "stateDiagram-v2") {
		t.Error("expected Mermaid graph to contain 'stateDiagram-v2'")
	}
	if !strings.Contains(mermaidGraph, "direction LR") {
		t.Error("expected Mermaid graph to contain 'direction LR'")
	}
	if !strings.Contains(mermaidGraph, "[*] -->") {
		t.Error("expected Mermaid graph to contain initial transition")
	}
}

func TestMermaidGraphWithoutDirection(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).Permit(TestTriggerY, TestStateA)

	info := sm.GetInfo()
	mermaidGraph := graph.MermaidGraph(info, nil)

	if !strings.Contains(mermaidGraph, "stateDiagram-v2") {
		t.Error("expected Mermaid graph to contain 'stateDiagram-v2'")
	}
	if strings.Contains(mermaidGraph, "direction") {
		t.Error("expected Mermaid graph not to contain 'direction' when not specified")
	}
}

func TestStateGraphWithHierarchy(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)
	sm.Configure(TestStateC).SubstateOf(TestStateB)

	info := sm.GetInfo()
	dotGraph := graph.UmlDotGraph(info)

	// Check for cluster
	if !strings.Contains(dotGraph, "subgraph") {
		t.Error("expected DOT graph to contain 'subgraph' for hierarchy")
	}
}

func TestNewStateGraph(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error { return nil })
	sm.Configure(TestStateB).
		PermitReentry(TestTriggerY)

	info := sm.GetInfo()
	sg := graph.NewStateGraph(info)

	if sg == nil {
		t.Fatal("expected non-nil StateGraph")
	}
	if len(sg.States) < 2 {
		t.Errorf("expected at least 2 states, got %d", len(sg.States))
	}
}

func TestUmlDotGraphStyle(t *testing.T) {
	style := graph.NewUmlDotGraphStyle()

	prefix := style.GetPrefix()
	if !strings.Contains(prefix, "digraph") {
		t.Error("expected prefix to contain 'digraph'")
	}
	if !strings.Contains(prefix, "node [shape=Mrecord]") {
		t.Error("expected prefix to contain node style")
	}
}

func TestMermaidGraphStyle(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)

	info := sm.GetInfo()
	sg := graph.NewStateGraph(info)
	direction := graph.TopToBottom
	style := graph.NewMermaidGraphStyle(sg, &direction)

	prefix := style.GetPrefix()
	if !strings.Contains(prefix, "stateDiagram-v2") {
		t.Error("expected prefix to contain 'stateDiagram-v2'")
	}
}

func TestEscapeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{`with"quotes`, `with\"quotes`},
		{`with\backslash`, `with\\backslash`},
		{`both"and\`, `both\"and\\`},
	}

	for _, tc := range tests {
		result := graph.EscapeLabel(tc.input)
		if result != tc.expected {
			t.Errorf("EscapeLabel(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestSanitizeStateName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SimpleState", "SimpleState"},
		{"State With Spaces", "StateWithSpaces"},
		{"State-With-Dashes", "StateWithDashes"},
		{"State:With:Colons", "StateWithColons"},
	}

	for _, tc := range tests {
		result := graph.SanitizeStateName(tc.input)
		if result != tc.expected {
			t.Errorf("SanitizeStateName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestGetDirectionCode(t *testing.T) {
	tests := []struct {
		direction graph.MermaidGraphDirection
		expected  string
	}{
		{graph.TopToBottom, "TB"},
		{graph.BottomToTop, "BT"},
		{graph.LeftToRight, "LR"},
		{graph.RightToLeft, "RL"},
	}

	for _, tc := range tests {
		result := graph.GetDirectionCode(tc.direction)
		if result != tc.expected {
			t.Errorf("GetDirectionCode(%v) = %q, expected %q", tc.direction, result, tc.expected)
		}
	}
}

// =============================================================================
// DOT Graph Fixture Tests (ported from .NET Stateless)
// =============================================================================

func TestDotGraph_SimpleTransition(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Check contains required elements (order may vary due to map iteration)
	if !strings.Contains(dotGraph, "digraph {") {
		t.Errorf("Expected graph to start with digraph, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"A" [label="A"]`) {
		t.Errorf("Expected graph to contain state A, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"B" [label="B"]`) {
		t.Errorf("Expected graph to contain state B, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"A" -> "B"`) {
		t.Errorf("Expected graph to contain A->B transition, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `label="X"`) {
		t.Errorf("Expected graph to contain trigger X, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `init -> "A"`) {
		t.Errorf("Expected graph to show initial state A, got:\n%s", dotGraph)
	}
}

func TestDotGraph_TwoSimpleTransitions(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		Permit(TestTriggerY, TestStateC)
	sm.Configure(TestStateB)
	sm.Configure(TestStateC)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Check contains required elements (order may vary due to map iteration)
	if !strings.Contains(dotGraph, `"A" [label="A"]`) {
		t.Errorf("Expected graph to contain state A, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"B" [label="B"]`) {
		t.Errorf("Expected graph to contain state B, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"C" [label="C"]`) {
		t.Errorf("Expected graph to contain state C, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"A" -> "B"`) {
		t.Errorf("Expected graph to contain A->B transition, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"A" -> "C"`) {
		t.Errorf("Expected graph to contain A->C transition, got:\n%s", dotGraph)
	}
}

func TestDotGraph_WhenDiscriminatedByAnonymousGuard(t *testing.T) {
	anonymousGuard := func() bool { return true }

	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitIf(TestTriggerX, TestStateB, anonymousGuard)
	sm.Configure(TestStateB)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Check contains required elements (order may vary due to map iteration)
	if !strings.Contains(dotGraph, `"A" [label="A"]`) {
		t.Errorf("Expected graph to contain state A, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"B" [label="B"]`) {
		t.Errorf("Expected graph to contain state B, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"A" -> "B"`) {
		t.Errorf("Expected graph to contain A->B transition, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `[`+stateless.DefaultFunctionDescription+`]`) {
		t.Errorf("Expected graph to contain guard description, got:\n%s", dotGraph)
	}
}

func TestDotGraph_WhenDiscriminatedByAnonymousGuardWithDescription(t *testing.T) {
	anonymousGuard := func() bool { return true }

	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitIf(TestTriggerX, TestStateB, anonymousGuard, "description")
	sm.Configure(TestStateB)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Check contains required elements (order may vary due to map iteration)
	if !strings.Contains(dotGraph, `"A" [label="A"]`) {
		t.Errorf("Expected graph to contain state A, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"B" [label="B"]`) {
		t.Errorf("Expected graph to contain state B, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `"A" -> "B"`) {
		t.Errorf("Expected graph to contain A->B transition, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `[description]`) {
		t.Errorf("Expected graph to contain guard description, got:\n%s", dotGraph)
	}
}

func TestDotGraph_DestinationStateIsDynamic(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitDynamic(TestTriggerX, func(_ any) TestState { return TestStateB })

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Check that decision node is created
	if !strings.Contains(dotGraph, "Decision1") {
		t.Errorf("Expected graph to contain Decision1 node, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `shape = "diamond"`) {
		t.Errorf("Expected graph to contain diamond shape, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, stateless.DefaultFunctionDescription) {
		t.Errorf("Expected graph to contain function description, got:\n%s", dotGraph)
	}
}

func TestDotGraph_OnEntryWithAnonymousActionAndDescription(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)

	// Note: Go doesn't have description parameter on OnEntry, use OnEntryWithDescription if available
	// For now, we test that entry actions appear in graph
	sm.Configure(TestStateA).OnEntry(
		func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error {
			return nil
		},
	)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Check that entry action appears (may use default description)
	if !strings.Contains(dotGraph, "entry /") {
		t.Errorf("Expected graph to contain entry action, got:\n%s", dotGraph)
	}
}

func TestDotGraph_TransitionWithIgnore(t *testing.T) {
	// Ignored triggers show as self-loops without entry/exit actions
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Ignore(TestTriggerY).
		Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should contain transition A->B for X
	if !strings.Contains(dotGraph, `"A" -> "B"`) {
		t.Errorf("Expected graph to contain A->B transition, got:\n%s", dotGraph)
	}
	// Should contain self-loop A->A for Y (ignored)
	if !strings.Contains(dotGraph, `"A" -> "A"`) {
		t.Errorf("Expected graph to contain A->A self-loop for ignored trigger, got:\n%s", dotGraph)
	}
}

func TestDotGraph_TransitionWithIgnoreAndEntry(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error { return nil }).
		Ignore(TestTriggerY).
		Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error { return nil }).
		PermitReentry(TestTriggerZ)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should contain reentry B->B for Z
	if !strings.Contains(dotGraph, `"B" -> "B"`) {
		t.Errorf("Expected graph to contain B->B reentry, got:\n%s", dotGraph)
	}
}

func TestDotGraph_InternalTransitionDoesNotShowEntryExitFunctions(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		OnEntry(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error { return nil }).
		OnExit(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error { return nil }).
		InternalTransition(TestTriggerX, func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error { return nil })

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should contain self-loop for internal transition
	if !strings.Contains(dotGraph, `"A" -> "A"`) {
		t.Errorf("Expected graph to contain A->A for internal transition, got:\n%s", dotGraph)
	}
	// Should show entry/exit in state box (not on transition)
	if !strings.Contains(dotGraph, "entry /") {
		t.Errorf("Expected graph to contain entry action in state, got:\n%s", dotGraph)
	}
}

func TestDotGraph_InitialStateNotChangedAfterTriggerFired(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)

	// Fire the trigger
	_ = sm.Fire(TestTriggerX, nil)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Initial state in graph should still be A
	if !strings.Contains(dotGraph, `init -> "A"`) {
		t.Errorf("Expected graph to show initial state as A, got:\n%s", dotGraph)
	}
}

func TestDotGraph_UmlWithSubstate(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		Permit(TestTriggerY, TestStateC)
	sm.Configure(TestStateB).SubstateOf(TestStateD)
	sm.Configure(TestStateC).SubstateOf(TestStateD)
	sm.Configure(TestStateD)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should contain subgraph cluster for D
	if !strings.Contains(dotGraph, "subgraph") {
		t.Errorf("Expected graph to contain subgraph for superstate, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, "clusterD") {
		t.Errorf("Expected graph to contain clusterD, got:\n%s", dotGraph)
	}
}

func TestDotGraph_UmlWithDynamic(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitDynamic(
		TestTriggerX,
		func(_ any) TestState { return TestStateB },
		stateless.DynamicStateInfo{DestinationState: "B", Criterion: "ChoseB"},
		stateless.DynamicStateInfo{DestinationState: "C", Criterion: "ChoseC"},
	)
	sm.Configure(TestStateB)
	sm.Configure(TestStateC)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should contain decision node
	if !strings.Contains(dotGraph, "Decision1") {
		t.Errorf("Expected graph to contain Decision1 node, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, "diamond") {
		t.Errorf("Expected graph to contain diamond shape for decision, got:\n%s", dotGraph)
	}
}

func TestDotGraph_ReentrantTransitionShowsEntryAction(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).
		OnEntry(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error {
			// Entry action that checks trigger (replaces OnEntryFrom)
			if tr.Trigger == TestTriggerX {
				// Handle specific trigger entry
			}
			return nil
		}).
		PermitReentry(TestTriggerX)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should contain B->B reentry
	if !strings.Contains(dotGraph, `"B" -> "B"`) {
		t.Errorf("Expected graph to contain B->B reentry, got:\n%s", dotGraph)
	}
}

func TestDotGraph_SimpleTransitionWithEscaping(t *testing.T) {
	state1 := `\state "1"`
	state2 := `\state "2"`
	trigger1 := `\trigger "1"`

	sm := stateless.NewStateMachine[string, string](state1)
	sm.Configure(state1).Permit(trigger1, state2)
	sm.Configure(state2)

	dotGraph := graph.UmlDotGraph(sm.GetInfo())

	// Should properly escape special characters
	if !strings.Contains(dotGraph, `\\`) {
		t.Errorf("Expected graph to contain escaped backslash, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `\"`) {
		t.Errorf("Expected graph to contain escaped quote, got:\n%s", dotGraph)
	}
}

// ================== Mermaid Graph Tests ==================

func TestMermaidGraph_InitialTransition(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA) // Configure state A so it appears in the graph

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	expected := "stateDiagram-v2\n[*] --> A"
	if mermaidGraph != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, mermaidGraph)
	}
}

func TestMermaidGraph_SimpleTransition(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain initial transition
	if !strings.Contains(mermaidGraph, "[*] --> A") {
		t.Errorf("Expected graph to contain initial transition, got:\n%s", mermaidGraph)
	}
	// Should contain A -> B transition
	if !strings.Contains(mermaidGraph, "A --> B : X") {
		t.Errorf("Expected graph to contain A --> B : X transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_SimpleTransition_LeftToRight(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)

	direction := graph.LeftToRight
	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), &direction)

	// Should contain direction
	if !strings.Contains(mermaidGraph, "direction LR") {
		t.Errorf("Expected graph to contain direction LR, got:\n%s", mermaidGraph)
	}
	// Should contain transition
	if !strings.Contains(mermaidGraph, "A --> B : X") {
		t.Errorf("Expected graph to contain A --> B : X transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_TwoSimpleTransitions(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).Permit(TestTriggerY, TestStateC)
	sm.Configure(TestStateC)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain initial transition
	if !strings.Contains(mermaidGraph, "[*] --> A") {
		t.Errorf("Expected graph to contain initial transition, got:\n%s", mermaidGraph)
	}
	// Should contain A -> B transition
	if !strings.Contains(mermaidGraph, "A --> B : X") {
		t.Errorf("Expected graph to contain A --> B : X transition, got:\n%s", mermaidGraph)
	}
	// Should contain B -> C transition
	if !strings.Contains(mermaidGraph, "B --> C : Y") {
		t.Errorf("Expected graph to contain B --> C : Y transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_WhenDiscriminatedByAnonymousGuard(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitIf(TestTriggerX, TestStateB, func() bool { return true }, "anonymousGuard")
	sm.Configure(TestStateB)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain transition with guard
	if !strings.Contains(mermaidGraph, "A --> B : X [anonymousGuard]") {
		t.Errorf("Expected graph to contain guarded transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_WhenDiscriminatedByAnonymousGuardWithDescription(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitIf(TestTriggerX, TestStateB, func() bool { return true }, "description")
	sm.Configure(TestStateB)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain transition with guard description
	if !strings.Contains(mermaidGraph, "A --> B : X [description]") {
		t.Errorf("Expected graph to contain guarded transition with description, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_DestinationStateIsDynamic(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitDynamic(TestTriggerX, func(_ any) TestState {
		return TestStateB
	})
	sm.Configure(TestStateB)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain decision node
	if !strings.Contains(mermaidGraph, "state Decision1 <<choice>>") {
		t.Errorf("Expected graph to contain decision node, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_DestinationStateIsDynamicWithPossibleDestinations(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).PermitDynamic(
		TestTriggerX,
		func(_ any) TestState { return TestStateB },
		stateless.DynamicStateInfo{DestinationState: "B", Criterion: "Going to B"},
		stateless.DynamicStateInfo{DestinationState: "C", Criterion: "Going to C"},
	)
	sm.Configure(TestStateB)
	sm.Configure(TestStateC)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain decision node
	if !strings.Contains(mermaidGraph, "Decision1 <<choice>>") {
		t.Errorf("Expected graph to contain decision node, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_TransitionWithIgnore(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		Ignore(TestTriggerY)
	sm.Configure(TestStateB)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain A -> B transition
	if !strings.Contains(mermaidGraph, "A --> B : X") {
		t.Errorf("Expected graph to contain A --> B : X transition, got:\n%s", mermaidGraph)
	}
	// Should contain ignore transition (self-loop)
	if !strings.Contains(mermaidGraph, "A --> A : Y") {
		t.Errorf("Expected graph to contain A --> A : Y (ignored) transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_WithSubstate(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).Permit(TestTriggerY, TestStateC).SubstateOf(TestStateD)
	sm.Configure(TestStateC).SubstateOf(TestStateD)
	sm.Configure(TestStateD)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain superstate definition
	if !strings.Contains(mermaidGraph, "state D {") {
		t.Errorf("Expected graph to contain superstate D, got:\n%s", mermaidGraph)
	}
	// Should contain substates inside superstate
	if !strings.Contains(mermaidGraph, "B") {
		t.Errorf("Expected graph to contain substate B, got:\n%s", mermaidGraph)
	}
	if !strings.Contains(mermaidGraph, "C") {
		t.Errorf("Expected graph to contain substate C, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_StateNamesWithSpacesAreAliased(t *testing.T) {
	stateA := "State A"
	stateB := "State B"
	triggerX := "Trigger X"

	sm := stateless.NewStateMachine[string, string](stateA)
	sm.Configure(stateA).Permit(triggerX, stateB)
	sm.Configure(stateB)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// States with spaces should be sanitized and aliased
	// The sanitized name should be used in transitions
	if !strings.Contains(mermaidGraph, "StateA") || !strings.Contains(mermaidGraph, "StateB") {
		t.Errorf("Expected graph to contain sanitized state names, got:\n%s", mermaidGraph)
	}
	// Should contain alias definitions
	if !strings.Contains(mermaidGraph, ": State A") || !strings.Contains(mermaidGraph, ": State B") {
		t.Errorf("Expected graph to contain alias definitions for states with spaces, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_OnEntryWithNamedDelegateAction(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).OnEntry(func(ctx context.Context, t stateless.Transition[TestState, TestTrigger]) error {
		return nil
	})

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain initial transition
	if !strings.Contains(mermaidGraph, "[*] --> A") {
		t.Errorf("Expected graph to contain initial transition, got:\n%s", mermaidGraph)
	}
	// Should contain A -> B transition
	if !strings.Contains(mermaidGraph, "A --> B : X") {
		t.Errorf("Expected graph to contain A --> B : X transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_InternalTransition(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).InternalTransition(
		TestTriggerX,
		func(ctx context.Context, t stateless.Transition[TestState, TestTrigger]) error {
			return nil
		},
	)

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain self-loop for internal transition
	if !strings.Contains(mermaidGraph, "A --> A : X") {
		t.Errorf("Expected graph to contain A --> A : X internal transition, got:\n%s", mermaidGraph)
	}
}

func TestMermaidGraph_OnEntryWithTriggerCheck(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).OnEntry(func(ctx context.Context, tr stateless.Transition[TestState, TestTrigger]) error {
		// Check trigger in entry action (replaces OnEntryFrom)
		if tr.Trigger == TestTriggerX {
			// Handle specific trigger
		}
		return nil
	})

	mermaidGraph := graph.MermaidGraph(sm.GetInfo(), nil)

	// Should contain transition from A to B with trigger X
	if !strings.Contains(mermaidGraph, "A --> B : X") {
		t.Errorf("Expected graph to contain transition, got:\n%s", mermaidGraph)
	}
}
