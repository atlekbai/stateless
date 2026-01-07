// Package graph provides visualization utilities for state machines.
package graph

import (
	"github.com/atlekbai/stateless"
)

// State represents a state in the graph.
type State struct {
	// StateName is the name of the state.
	StateName string

	// NodeName is the name used for the node in the graph.
	NodeName string

	// EntryActions are the entry actions for this state.
	EntryActions []string

	// ExitActions are the exit actions for this state.
	ExitActions []string

	// Leaving are the transitions leaving this state.
	Leaving []*Transition

	// Arriving are the transitions arriving at this state.
	Arriving []*Transition

	// SuperState is the parent state, if any.
	SuperState *SuperState

	// StateInfo contains the underlying state information.
	StateInfo *stateless.StateInfo
}

// SuperState represents a state that contains substates.
type SuperState struct {
	*State

	// SubStates are the child states of this state.
	SubStates []*State
}

// Decision represents a decision node in the graph (for dynamic transitions).
type Decision struct {
	// NodeName is the name of the decision node.
	NodeName string

	// Method contains information about the decision method.
	Method stateless.InvocationInfo

	// Leaving are the transitions leaving this decision node.
	Leaving []*Transition

	// Arriving are the transitions arriving at this decision node.
	Arriving []*Transition
}

// Transition represents a transition in the graph.
type Transition struct {
	// Trigger is the trigger that causes this transition.
	Trigger stateless.TriggerInfo

	// SourceState is the source state of the transition.
	SourceState *State

	// DestinationState is the destination state of the transition.
	DestinationState *State

	// Guards are the guard conditions for this transition.
	Guards []stateless.InvocationInfo

	// DestinationEntryActions are the entry actions executed at the destination.
	DestinationEntryActions []stateless.ActionInfo

	// ExecuteEntryExitActions indicates if entry/exit actions should be executed.
	ExecuteEntryExitActions bool
}

// StayTransition represents a transition from a state to itself.
type StayTransition struct {
	*Transition
}

// FixedTransition represents a transition to a fixed destination state.
type FixedTransition struct {
	*Transition
}

// DynamicTransition represents a transition to a dynamically determined state.
type DynamicTransition struct {
	*Transition

	// Criterion is the reason this destination was chosen.
	Criterion string
}
