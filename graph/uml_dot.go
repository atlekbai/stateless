package graph

import (
	"fmt"
	"strings"

	"github.com/atlekbai/stateless"
)

// UmlDotGraphStyle generates DOT graphs in basic UML style.
type UmlDotGraphStyle struct{}

// NewUmlDotGraphStyle creates a new UML DOT graph style.
func NewUmlDotGraphStyle() *UmlDotGraphStyle {
	return &UmlDotGraphStyle{}
}

// GetPrefix returns the text that starts a new DOT graph.
func (s *UmlDotGraphStyle) GetPrefix() string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString("compound=true;\n")
	sb.WriteString("node [shape=Mrecord]\n")
	sb.WriteString("rankdir=\"LR\"\n")
	return sb.String()
}

// FormatOneCluster formats a superstate and its substates.
func (s *UmlDotGraphStyle) FormatOneCluster(superState *SuperState) string {
	var sb strings.Builder
	var label strings.Builder

	label.WriteString(EscapeLabel(superState.StateName))

	if len(superState.EntryActions) > 0 || len(superState.ExitActions) > 0 {
		label.WriteString("\\n----------")
		for _, act := range superState.EntryActions {
			label.WriteString("\\nentry / ")
			label.WriteString(EscapeLabel(act))
		}
		for _, act := range superState.ExitActions {
			label.WriteString("\\nexit / ")
			label.WriteString(EscapeLabel(act))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("subgraph \"cluster%s\"\n", EscapeLabel(superState.NodeName)))
	sb.WriteString("\t{\n")
	sb.WriteString(fmt.Sprintf("\tlabel = \"%s\"\n", label.String()))

	for _, subState := range superState.SubStates {
		sb.WriteString(s.FormatOneState(subState))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// FormatOneState formats a single state.
func (s *UmlDotGraphStyle) FormatOneState(state *State) string {
	escapedName := EscapeLabel(state.StateName)

	if len(state.EntryActions) == 0 && len(state.ExitActions) == 0 {
		return fmt.Sprintf("\"%s\" [label=\"%s\"];\n", escapedName, escapedName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\"%s\" [label=\"%s|", escapedName, escapedName))

	var actions []string
	for _, act := range state.EntryActions {
		actions = append(actions, "entry / "+EscapeLabel(act))
	}
	for _, act := range state.ExitActions {
		actions = append(actions, "exit / "+EscapeLabel(act))
	}

	sb.WriteString(strings.Join(actions, "\\n"))
	sb.WriteString("\"];\n")

	return sb.String()
}

// FormatOneDecisionNode formats a decision node.
func (s *UmlDotGraphStyle) FormatOneDecisionNode(nodeName, label string) string {
	return fmt.Sprintf("\"%s\" [shape = \"diamond\", label = \"%s\"];\n",
		EscapeLabel(nodeName), EscapeLabel(label))
}

// FormatAllTransitions formats all transitions.
func (s *UmlDotGraphStyle) FormatAllTransitions(
	transitions []*Transition,
	_ []*Decision,
) []string {
	return FormatTransitions(s, transitions)
}

// FormatOneTransition formats a single transition.
func (s *UmlDotGraphStyle) FormatOneTransition(
	sourceNodeName, trigger string,
	actions []string,
	destinationNodeName string,
	guards []string,
) string {
	var sb strings.Builder

	sb.WriteString(trigger)

	if len(actions) > 0 {
		sb.WriteString(" / ")
		sb.WriteString(strings.Join(actions, ", "))
	}

	if len(guards) > 0 {
		for _, info := range guards {
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString("[")
			sb.WriteString(info)
			sb.WriteString("]")
		}
	}

	return formatOneLine(sourceNodeName, destinationNodeName, sb.String())
}

// GetInitialTransition returns the text for the initial state transition.
func (s *UmlDotGraphStyle) GetInitialTransition(initialState *stateless.StateInfo) string {
	if initialState == nil {
		return "\n}"
	}

	initialStateName := fmt.Sprintf("%v", initialState.UnderlyingState)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(" init [label=\"\", shape=point];")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(" init -> \"%s\"[style = \"solid\"]", EscapeLabel(initialStateName)))
	sb.WriteString("\n")
	sb.WriteString("}")

	return sb.String()
}

// formatOneLine formats a single transition line.
func formatOneLine(fromNodeName, toNodeName, label string) string {
	return fmt.Sprintf("\"%s\" -> \"%s\" [style=\"solid\", label=\"%s\"];",
		EscapeLabel(fromNodeName), EscapeLabel(toNodeName), EscapeLabel(label))
}

// EscapeLabel escapes special characters in a label.
func EscapeLabel(label string) string {
	label = strings.ReplaceAll(label, "\\", "\\\\")
	label = strings.ReplaceAll(label, "\"", "\\\"")
	return label
}

// UmlDotGraph generates a UML DOT graph from state machine info.
func UmlDotGraph(machineInfo *stateless.StateMachineInfo) string {
	graph := NewStateGraph(machineInfo)
	return graph.ToGraph(NewUmlDotGraphStyle())
}
