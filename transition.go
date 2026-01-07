package stateless

// Transition describes a state transition.
type Transition[TState, TTrigger comparable] struct {
	// Source is the state transitioned from.
	Source TState

	// Destination is the state transitioned to.
	Destination TState

	// Trigger is the trigger that caused the transition.
	Trigger TTrigger

	// Parameters are the optional trigger parameters.
	Parameters []any

	// isInitial indicates if this is an initial transition (entering the state machine).
	isInitial bool
}

// NewTransition creates a new transition.
func NewTransition[TState, TTrigger comparable](source, destination TState, trigger TTrigger, parameters ...any) Transition[TState, TTrigger] {
	if parameters == nil {
		parameters = []any{}
	}
	return Transition[TState, TTrigger]{
		Source:      source,
		Destination: destination,
		Trigger:     trigger,
		Parameters:  parameters,
	}
}

// NewInitialTransition creates a new initial transition.
func NewInitialTransition[TState, TTrigger comparable](source, destination TState, trigger TTrigger, parameters ...any) Transition[TState, TTrigger] {
	t := NewTransition(source, destination, trigger, parameters...)
	t.isInitial = true
	return t
}

// IsReentry returns true if the transition is a re-entry, i.e., the identity transition.
func (t Transition[TState, TTrigger]) IsReentry() bool {
	return any(t.Source) == any(t.Destination)
}

// IsInitial returns true if this is an initial transition.
func (t Transition[TState, TTrigger]) IsInitial() bool {
	return t.isInitial
}
