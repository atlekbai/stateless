package stateless

import "context"

// TransitionAction is a function that is executed during a state transition.
// It receives a context and the transition information, and returns an error if the action fails.
type TransitionAction[TState, TTrigger comparable] func(
	ctx context.Context,
	t Transition[TState, TTrigger],
) error

// Transition describes a state transition.
type Transition[TState, TTrigger comparable] struct {
	// Source is the state transitioned from.
	Source TState

	// Destination is the state transitioned to.
	Destination TState

	// Trigger is the trigger that caused the transition.
	Trigger TTrigger

	// Args contains the arguments passed with the trigger.
	// Use type assertion to access typed arguments:
	//   if args, ok := t.Args.(MyArgs); ok { ... }
	Args any

	// isInitial indicates if this is an initial transition (entering the state machine).
	isInitial bool
}

// NewTransition creates a new transition.
func NewTransition[TState, TTrigger comparable](
	source, destination TState,
	trigger TTrigger,
	args any,
) Transition[TState, TTrigger] {
	return Transition[TState, TTrigger]{
		Source:      source,
		Destination: destination,
		Trigger:     trigger,
		Args:        args,
	}
}

// NewInitialTransition creates a new initial transition.
func NewInitialTransition[TState, TTrigger comparable](
	source, destination TState,
	trigger TTrigger,
	args any,
) Transition[TState, TTrigger] {
	return Transition[TState, TTrigger]{
		Source:      source,
		Destination: destination,
		Trigger:     trigger,
		Args:        args,
		isInitial:   true,
	}
}

// IsReentry returns true if the transition is a re-entry, i.e., the identity transition.
func (t Transition[TState, TTrigger]) IsReentry() bool {
	return any(t.Source) == any(t.Destination)
}

// IsInitial returns true if this is an initial transition.
func (t Transition[TState, TTrigger]) IsInitial() bool {
	return t.isInitial
}
