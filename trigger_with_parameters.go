package stateless

// TriggerDetails represents a trigger with details about its configuration.
type TriggerDetails[TState, TTrigger comparable] struct {
	// Trigger is the trigger value.
	Trigger TTrigger
}

// NewTriggerDetails creates a new TriggerDetails.
func NewTriggerDetails[TState, TTrigger comparable](trigger TTrigger) TriggerDetails[TState, TTrigger] {
	return TriggerDetails[TState, TTrigger]{
		Trigger: trigger,
	}
}
