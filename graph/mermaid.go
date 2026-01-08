package graph

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/atlekbai/stateless"
)

// MermaidGraphDirection specifies the direction of the Mermaid graph.
type MermaidGraphDirection int

const (
	// TopToBottom flows from top to bottom.
	TopToBottom MermaidGraphDirection = iota
	// BottomToTop flows from bottom to top.
	BottomToTop
	// LeftToRight flows from left to right.
	LeftToRight
	// RightToLeft flows from right to left.
	RightToLeft
)

// MermaidGraphStyle generates Mermaid graphs.
type MermaidGraphStyle struct {
	graph               *StateGraph
	direction           *MermaidGraphDirection
	stateMap            map[string]*State
	stateMapInitialized bool
}

// NewMermaidGraphStyle creates a new Mermaid graph style.
func NewMermaidGraphStyle(graph *StateGraph, direction *MermaidGraphDirection) *MermaidGraphStyle {
	return &MermaidGraphStyle{
		graph:     graph,
		direction: direction,
		stateMap:  make(map[string]*State),
	}
}

// GetPrefix returns the text that starts a new Mermaid graph.
func (s *MermaidGraphStyle) GetPrefix() string {
	s.buildSanitizedNamedStateMap()

	var sb strings.Builder
	sb.WriteString("stateDiagram-v2")

	if s.direction != nil {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("\tdirection %s", getDirectionCode(*s.direction)))
	}

	// Add state aliases for states with sanitized names
	for sanitizedName, state := range s.stateMap {
		if sanitizedName != state.StateName {
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("\t%s : %s", sanitizedName, state.StateName))
		}
	}

	return sb.String()
}

// FormatOneCluster formats a superstate and its substates.
func (s *MermaidGraphStyle) FormatOneCluster(superState *SuperState) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("\tstate %s {\n", s.getSanitizedStateName(superState.StateName)))

	for _, subState := range superState.SubStates {
		sb.WriteString(fmt.Sprintf("\t\t%s\n", s.getSanitizedStateName(subState.StateName)))
	}

	sb.WriteString("\t}")
	return sb.String()
}

// FormatOneState formats a single state (Mermaid doesn't need explicit state definitions).
func (s *MermaidGraphStyle) FormatOneState(_ *State) string {
	return ""
}

// FormatOneDecisionNode formats a decision node.
func (s *MermaidGraphStyle) FormatOneDecisionNode(nodeName, _ string) string {
	return fmt.Sprintf("\n\tstate %s <<choice>>", nodeName)
}

// FormatAllTransitions formats all transitions.
func (s *MermaidGraphStyle) FormatAllTransitions(
	transitions []*Transition,
	_ []*Decision,
) []string {
	return FormatTransitions(s, transitions)
}

// FormatOneTransition formats a single transition.
func (s *MermaidGraphStyle) FormatOneTransition(
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

	sanitizedSource := s.getSanitizedStateName(sourceNodeName)
	sanitizedDest := s.getSanitizedStateName(destinationNodeName)

	return fmt.Sprintf("\t%s --> %s : %s", sanitizedSource, sanitizedDest, sb.String())
}

// GetInitialTransition returns the text for the initial state transition.
func (s *MermaidGraphStyle) GetInitialTransition(initialState *stateless.StateInfo) string {
	if initialState == nil {
		return ""
	}

	sanitizedStateName := s.getSanitizedStateName(fmt.Sprintf("%v", initialState.UnderlyingState))
	return fmt.Sprintf("\n[*] --> %s", sanitizedStateName)
}

// buildSanitizedNamedStateMap builds a map of sanitized state names to states.
func (s *MermaidGraphStyle) buildSanitizedNamedStateMap() {
	if s.stateMapInitialized {
		return
	}

	uniqueAliases := make(map[string]bool)

	for _, state := range s.graph.States {
		sanitizedName := sanitizeStateName(state.StateName)

		if sanitizedName != state.StateName {
			count := 1
			tempName := sanitizedName
			for uniqueAliases[tempName] || s.graph.States[tempName] != nil {
				tempName = fmt.Sprintf("%s_%d", sanitizedName, count)
				count++
			}
			sanitizedName = tempName
			uniqueAliases[sanitizedName] = true
		}

		s.stateMap[sanitizedName] = state
	}

	s.stateMapInitialized = true
}

// getSanitizedStateName returns the sanitized name for a state.
func (s *MermaidGraphStyle) getSanitizedStateName(stateName string) string {
	for sanitizedName, state := range s.stateMap {
		if state.StateName == stateName {
			return sanitizedName
		}
	}
	return stateName
}

// sanitizeStateName removes characters that would cause invalid Mermaid graphs.
func sanitizeStateName(name string) string {
	var result strings.Builder
	for _, c := range name {
		if !unicode.IsSpace(c) && c != ':' && c != '-' {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// getDirectionCode returns the Mermaid direction code.
func getDirectionCode(direction MermaidGraphDirection) string {
	switch direction {
	case TopToBottom:
		return "TB"
	case BottomToTop:
		return "BT"
	case LeftToRight:
		return "LR"
	case RightToLeft:
		return "RL"
	default:
		return "TB"
	}
}

// MermaidGraph generates a Mermaid graph from state machine info.
func MermaidGraph(machineInfo *stateless.StateMachineInfo, direction *MermaidGraphDirection) string {
	graph := NewStateGraph(machineInfo)
	return graph.ToGraph(NewMermaidGraphStyle(graph, direction))
}
