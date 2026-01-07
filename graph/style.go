package graph

import (
	"github.com/atlekbai/stateless"
)

// GraphStyle defines the interface for formatting state graphs.
type GraphStyle interface {
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
	FormatOneTransition(sourceNodeName, trigger string, actions []string, destinationNodeName string, guards []string) string
}
