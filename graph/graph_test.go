package graph

import (
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
)

const (
	TestTriggerX TestTrigger = iota
	TestTriggerY
	TestTriggerZ
)

func (s TestState) String() string {
	switch s {
	case TestStateA:
		return "TestStateA"
	case TestStateB:
		return "TestStateB"
	case TestStateC:
		return "TestStateC"
	default:
		return "Unknown"
	}
}

func (t TestTrigger) String() string {
	switch t {
	case TestTriggerX:
		return "TestTriggerX"
	case TestTriggerY:
		return "TestTriggerY"
	case TestTriggerZ:
		return "TestTriggerZ"
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
	if !strings.Contains(graph, "TestStateA") {
		t.Error("expected DOT graph to contain 'TestStateA'")
	}
	if !strings.Contains(graph, "TestStateB") {
		t.Error("expected DOT graph to contain 'TestStateB'")
	}
	if !strings.Contains(graph, "TestStateC") {
		t.Error("expected DOT graph to contain 'TestStateC'")
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
