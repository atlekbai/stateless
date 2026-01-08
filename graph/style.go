package graph

import (
	"fmt"

	"github.com/atlekbai/stateless"
)

// Style defines the interface for formatting state graphs.
type Style interface {
	// GetPrefix returns the text that starts a new graph.
	GetPrefix() string

	// GetInitialTransition returns the text for the initial state transition.
	GetInitialTransition(initialState *stateless.StateInfo) string

	// FormatOneState formats a single state.
	FormatOneState(state *State) string

	// FormatOneCluster formats a superstate and its substates.
	FormatOneCluster(superState *SuperState) string

	// FormatOneDecisionNode formats a decision node.
	FormatOneDecisionNode(nodeName, label string) string

	// FormatAllTransitions formats all transitions.
	FormatAllTransitions(transitions []*Transition, decisions []*Decision) []string

	// FormatOneTransition formats a single transition.
	FormatOneTransition(
		sourceNodeName, trigger string,
		actions []string,
		destinationNodeName string,
		guards []string,
	) string
}

// FormatTransitions is a helper that formats all transitions using the given style.
// This eliminates duplicate logic between different style implementations.
func FormatTransitions(style Style, transitions []*Transition) []string {
	var lines []string

	for _, transit := range transitions {
		line := formatSingleTransition(style, transit)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func formatSingleTransition(style Style, transit *Transition) string {
	// Determine if this is a stay transition
	if transit.SourceState == transit.DestinationState {
		return formatStayTransition(style, transit)
	} else if transit.DestinationState != nil {
		return formatRegularTransition(style, transit)
	}
	return ""
}

func formatStayTransition(style Style, transit *Transition) string {
	var actions []string
	if transit.ExecuteEntryExitActions {
		for _, act := range transit.DestinationEntryActions {
			actions = append(actions, act.Description())
		}
	}

	guards := collectGuards(transit)

	if !transit.ExecuteEntryExitActions {
		actions = nil
	}

	return style.FormatOneTransition(
		transit.SourceState.NodeName,
		fmt.Sprintf("%v", transit.Trigger.UnderlyingTrigger),
		actions,
		transit.SourceState.NodeName,
		guards,
	)
}

func formatRegularTransition(style Style, transit *Transition) string {
	var actions []string
	for _, act := range transit.DestinationEntryActions {
		actions = append(actions, act.Description())
	}

	guards := collectGuards(transit)

	return style.FormatOneTransition(
		transit.SourceState.NodeName,
		fmt.Sprintf("%v", transit.Trigger.UnderlyingTrigger),
		actions,
		transit.DestinationState.NodeName,
		guards,
	)
}

func collectGuards(transit *Transition) []string {
	var guards []string
	for _, g := range transit.Guards {
		guards = append(guards, g.Description())
	}
	return guards
}
