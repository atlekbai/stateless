package stateless

// NoArgs is used when a transition has no arguments.
type NoArgs struct{}

// Transition describes a state transition with typed arguments.
type Transition[TState, TTrigger comparable, TArgs any] struct {
	// Source is the state transitioned from.
	Source TState

	// Destination is the state transitioned to.
	Destination TState

	// Trigger is the trigger that caused the transition.
	Trigger TTrigger

	// Args contains the typed arguments passed with the trigger.
	Args TArgs

	// isInitial indicates if this is an initial transition (entering the state machine).
	isInitial bool
}

// NewTransition creates a new transition with typed arguments.
func NewTransition[TState, TTrigger comparable, TArgs any](source, destination TState, trigger TTrigger, args TArgs) Transition[TState, TTrigger, TArgs] {
	return Transition[TState, TTrigger, TArgs]{
		Source:      source,
		Destination: destination,
		Trigger:     trigger,
		Args:        args,
	}
}

// NewInitialTransition creates a new initial transition.
func NewInitialTransition[TState, TTrigger comparable, TArgs any](source, destination TState, trigger TTrigger, args TArgs) Transition[TState, TTrigger, TArgs] {
	return Transition[TState, TTrigger, TArgs]{
		Source:      source,
		Destination: destination,
		Trigger:     trigger,
		Args:        args,
		isInitial:   true,
	}
}

// IsReentry returns true if the transition is a re-entry, i.e., the identity transition.
func (t Transition[TState, TTrigger, TArgs]) IsReentry() bool {
	return any(t.Source) == any(t.Destination)
}

// IsInitial returns true if this is an initial transition.
func (t Transition[TState, TTrigger, TArgs]) IsInitial() bool {
	return t.isInitial
}

// internalTransition is used internally with untyped args for storage.
type internalTransition[TState, TTrigger comparable] struct {
	Source      TState
	Destination TState
	Trigger     TTrigger
	Args        any
	isInitial   bool
}

// toTyped converts an internal transition to a typed transition.
func toTypedTransition[TState, TTrigger comparable, TArgs any](t internalTransition[TState, TTrigger]) Transition[TState, TTrigger, TArgs] {
	var args TArgs
	if t.Args != nil {
		args, _ = t.Args.(TArgs)
	}
	return Transition[TState, TTrigger, TArgs]{
		Source:      t.Source,
		Destination: t.Destination,
		Trigger:     t.Trigger,
		Args:        args,
		isInitial:   t.isInitial,
	}
}
