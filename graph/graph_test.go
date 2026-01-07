package graph

import (
	"fmt"
	"strings"
	"testing"

	"github.com/atlekbai/stateless"
)

// Test state and trigger types
type TestState int
type TestTrigger int

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
	graph := UmlDotGraph(info)

	// Check basic structure
	if !strings.Contains(graph, "digraph") {
		t.Error("expected DOT graph to contain 'digraph'")
	}
	if !strings.Contains(graph, "init") {
		t.Error("expected DOT graph to contain 'init' node")
	}
	if !strings.Contains(graph, "\"A\"") {
		t.Error("expected DOT graph to contain 'A'")
	}
	if !strings.Contains(graph, "\"B\"") {
		t.Error("expected DOT graph to contain 'B'")
	}
	if !strings.Contains(graph, "\"C\"") {
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
	direction := LeftToRight
	graph := MermaidGraph(info, &direction)

	// Check basic structure
	if !strings.Contains(graph, "stateDiagram-v2") {
		t.Error("expected Mermaid graph to contain 'stateDiagram-v2'")
	}
	if !strings.Contains(graph, "direction LR") {
		t.Error("expected Mermaid graph to contain 'direction LR'")
	}
	if !strings.Contains(graph, "[*] -->") {
		t.Error("expected Mermaid graph to contain initial transition")
	}
}

func TestMermaidGraphWithoutDirection(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).Permit(TestTriggerY, TestStateA)

	info := sm.GetInfo()
	graph := MermaidGraph(info, nil)

	if !strings.Contains(graph, "stateDiagram-v2") {
		t.Error("expected Mermaid graph to contain 'stateDiagram-v2'")
	}
	if strings.Contains(graph, "direction") {
		t.Error("expected Mermaid graph not to contain 'direction' when not specified")
	}
}

func TestStateGraphWithHierarchy(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)
	sm.Configure(TestStateC).SubstateOf(TestStateB)

	info := sm.GetInfo()
	graph := UmlDotGraph(info)

	// Check for cluster
	if !strings.Contains(graph, "subgraph") {
		t.Error("expected DOT graph to contain 'subgraph' for hierarchy")
	}
}

func TestNewStateGraph(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		Permit(TestTriggerX, TestStateB).
		OnEntry(func() {})
	sm.Configure(TestStateB).
		PermitReentry(TestTriggerY)

	info := sm.GetInfo()
	sg := NewStateGraph(info)

	if sg == nil {
		t.Fatal("expected non-nil StateGraph")
	}
	if len(sg.States) < 2 {
		t.Errorf("expected at least 2 states, got %d", len(sg.States))
	}
}

func TestUmlDotGraphStyle(t *testing.T) {
	style := NewUmlDotGraphStyle()

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
	sg := NewStateGraph(info)
	direction := TopToBottom
	style := NewMermaidGraphStyle(sg, &direction)

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
		result := escapeLabel(tc.input)
		if result != tc.expected {
			t.Errorf("escapeLabel(%q) = %q, expected %q", tc.input, result, tc.expected)
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
		result := sanitizeStateName(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeStateName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestGetDirectionCode(t *testing.T) {
	tests := []struct {
		direction MermaidGraphDirection
		expected  string
	}{
		{TopToBottom, "TB"},
		{BottomToTop, "BT"},
		{LeftToRight, "LR"},
		{RightToLeft, "RL"},
	}

	for _, tc := range tests {
		result := getDirectionCode(tc.direction)
		if result != tc.expected {
			t.Errorf("getDirectionCode(%v) = %q, expected %q", tc.direction, result, tc.expected)
		}
	}
}

// =============================================================================
// DOT Graph Fixture Tests (ported from .NET Stateless)
// =============================================================================

// Helper functions for building expected DOT graph strings

func dotPrefix() string {
	return "digraph {\n" +
		"compound=true;\n" +
		"node [shape=Mrecord]\n" +
		"rankdir=\"LR\"\n"
}

func dotSuffix(initialState string) string {
	return "\n" +
		" init [label=\"\", shape=point];\n" +
		fmt.Sprintf(" init -> \"%s\"[style = \"solid\"]\n", escapeLabel(initialState)) +
		"}"
}

func dotBox(label string, entries []string, exits []string) string {
	var es []string
	for _, entry := range entries {
		es = append(es, "entry / "+entry)
	}
	for _, exit := range exits {
		es = append(es, "exit / "+exit)
	}

	if len(es) == 0 {
		return fmt.Sprintf("\"%s\" [label=\"%s\"];\n", label, label)
	}
	return fmt.Sprintf("\"%s\" [label=\"%s|%s\"];\n", label, label, strings.Join(es, "\\n"))
}

func dotDecision(nodeName, label string) string {
	return fmt.Sprintf("\"%s\" [shape = \"diamond\", label = \"%s\"];\n", nodeName, label)
}

func dotLine(from, to string, label *string) string {
	s := fmt.Sprintf("\"%s\" -> \"%s\" [style=\"solid\"", from, to)
	if label != nil {
		s += fmt.Sprintf(", label=\"%s\"", *label)
	}
	s += "];"
	return s
}

func dotSubgraph(graphName, label, contents string) string {
	return "\n" +
		fmt.Sprintf("subgraph \"cluster%s\"\n", graphName) +
		"\t{\n" +
		fmt.Sprintf("\tlabel = \"%s\"\n", label) +
		contents +
		"}\n"
}

func strPtr(s string) *string {
	return &s
}

func TestDotGraph_SimpleTransition(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB)

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

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
	sm.Configure(TestStateA).PermitDynamic(TestTriggerX, func() TestState { return TestStateB })

	dotGraph := UmlDotGraph(sm.GetInfo())

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
	sm.Configure(TestStateA).OnEntry(func() {})

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

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
		OnEntry(func() {}).
		Ignore(TestTriggerY).
		Permit(TestTriggerX, TestStateB)
	sm.Configure(TestStateB).
		OnEntry(func() {}).
		PermitReentry(TestTriggerZ)

	dotGraph := UmlDotGraph(sm.GetInfo())

	// Should contain reentry B->B for Z
	if !strings.Contains(dotGraph, `"B" -> "B"`) {
		t.Errorf("Expected graph to contain B->B reentry, got:\n%s", dotGraph)
	}
}

func TestDotGraph_InternalTransitionDoesNotShowEntryExitFunctions(t *testing.T) {
	sm := stateless.NewStateMachine[TestState, TestTrigger](TestStateA)
	sm.Configure(TestStateA).
		OnEntry(func() {}).
		OnExit(func() {}).
		InternalTransition(TestTriggerX, func() {})

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

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
		func() TestState { return TestStateB },
		stateless.DynamicStateInfo{DestinationState: "B", Criterion: "ChoseB"},
		stateless.DynamicStateInfo{DestinationState: "C", Criterion: "ChoseC"},
	)
	sm.Configure(TestStateB)
	sm.Configure(TestStateC)

	dotGraph := UmlDotGraph(sm.GetInfo())

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
		OnEntryFrom(TestTriggerX, func() {}).
		PermitReentry(TestTriggerX)

	dotGraph := UmlDotGraph(sm.GetInfo())

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

	dotGraph := UmlDotGraph(sm.GetInfo())

	// Should properly escape special characters
	if !strings.Contains(dotGraph, `\\`) {
		t.Errorf("Expected graph to contain escaped backslash, got:\n%s", dotGraph)
	}
	if !strings.Contains(dotGraph, `\"`) {
		t.Errorf("Expected graph to contain escaped quote, got:\n%s", dotGraph)
	}
}
