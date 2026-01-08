package stateless

import "context"

// EntryActionBehaviour represents an entry action for a state.
type EntryActionBehaviour[TState, TTrigger comparable] struct {
	action      TransitionAction[TState, TTrigger]
	description InvocationInfo
}

// NewEntryActionBehaviour creates a new entry action behaviour.
func NewEntryActionBehaviour[TState, TTrigger comparable](
	action TransitionAction[TState, TTrigger],
	description InvocationInfo,
) *EntryActionBehaviour[TState, TTrigger] {
	return &EntryActionBehaviour[TState, TTrigger]{
		action:      action,
		description: description,
	}
}

// Execute executes the entry action.
func (s *EntryActionBehaviour[TState, TTrigger]) Execute(
	ctx context.Context,
	transition Transition[TState, TTrigger],
) error {
	if s.action != nil {
		return s.action(ctx, transition)
	}
	return nil
}

// GetDescription returns the description of the action.
func (s *EntryActionBehaviour[TState, TTrigger]) GetDescription() InvocationInfo {
	return s.description
}

// ExitActionBehaviour represents an exit action for a state.
type ExitActionBehaviour[TState, TTrigger comparable] struct {
	action      TransitionAction[TState, TTrigger]
	description InvocationInfo
}

// NewExitActionBehaviour creates a new exit action behaviour.
func NewExitActionBehaviour[TState, TTrigger comparable](
	action TransitionAction[TState, TTrigger],
	description InvocationInfo,
) *ExitActionBehaviour[TState, TTrigger] {
	return &ExitActionBehaviour[TState, TTrigger]{
		action:      action,
		description: description,
	}
}

// Execute executes the exit action.
func (s *ExitActionBehaviour[TState, TTrigger]) Execute(ctx context.Context, t Transition[TState, TTrigger]) error {
	if s.action != nil {
		return s.action(ctx, t)
	}
	return nil
}

// GetDescription returns the description of the action.
func (s *ExitActionBehaviour[TState, TTrigger]) GetDescription() InvocationInfo {
	return s.description
}

// ActivateActionBehaviour represents an activation action for a state.
type ActivateActionBehaviour[TState comparable] struct {
	action      func(ctx context.Context) error
	description InvocationInfo
}

// NewActivateActionBehaviour creates a new activation action behaviour.
func NewActivateActionBehaviour[TState comparable](
	action func(ctx context.Context) error,
	description InvocationInfo,
) *ActivateActionBehaviour[TState] {
	return &ActivateActionBehaviour[TState]{
		action:      action,
		description: description,
	}
}

// Execute executes the activation action.
func (s *ActivateActionBehaviour[TState]) Execute(ctx context.Context) error {
	if s.action != nil {
		return s.action(ctx)
	}
	return nil
}

// GetDescription returns the description of the action.
func (s *ActivateActionBehaviour[TState]) GetDescription() InvocationInfo {
	return s.description
}

// DeactivateActionBehaviour represents a deactivation action for a state.
type DeactivateActionBehaviour[TState comparable] struct {
	action      func(ctx context.Context) error
	description InvocationInfo
}

// NewDeactivateActionBehaviour creates a new deactivation action behaviour.
func NewDeactivateActionBehaviour[TState comparable](
	action func(ctx context.Context) error,
	description InvocationInfo,
) *DeactivateActionBehaviour[TState] {
	return &DeactivateActionBehaviour[TState]{
		action:      action,
		description: description,
	}
}

// Execute executes the deactivation action.
func (s *DeactivateActionBehaviour[TState]) Execute(ctx context.Context) error {
	if s.action != nil {
		return s.action(ctx)
	}
	return nil
}

// GetDescription returns the description of the action.
func (s *DeactivateActionBehaviour[TState]) GetDescription() InvocationInfo {
	return s.description
}
