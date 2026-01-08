package stateless

import "context"

// EntryActionBehaviour represents an entry action for a state.
type EntryActionBehaviour[TState, TTrigger comparable] interface {
	// Execute executes the entry action.
	Execute(ctx context.Context, transition Transition[TState, TTrigger]) error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
}

// SyncEntryActionBehaviour is a synchronous entry action.
type SyncEntryActionBehaviour[TState, TTrigger comparable] struct {
	action      TransitionAction[TState, TTrigger]
	description InvocationInfo
}

// NewSyncEntryActionBehaviour creates a new synchronous entry action.
func NewSyncEntryActionBehaviour[TState, TTrigger comparable](
	action TransitionAction[TState, TTrigger],
	description InvocationInfo,
) *SyncEntryActionBehaviour[TState, TTrigger] {
	return &SyncEntryActionBehaviour[TState, TTrigger]{
		action:      action,
		description: description,
	}
}

func (s *SyncEntryActionBehaviour[TState, TTrigger]) Execute(
	ctx context.Context,
	transition Transition[TState, TTrigger],
) error {
	if s.action != nil {
		return s.action(ctx, transition)
	}
	return nil
}

func (s *SyncEntryActionBehaviour[TState, TTrigger]) GetDescription() InvocationInfo {
	return s.description
}

// ExitActionBehaviour represents an exit action for a state.
type ExitActionBehaviour[TState, TTrigger comparable] interface {
	// Execute executes the exit action.
	Execute(ctx context.Context, transition Transition[TState, TTrigger]) error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
}

// SyncExitActionBehaviour is a synchronous exit action.
type SyncExitActionBehaviour[TState, TTrigger comparable] struct {
	action      TransitionAction[TState, TTrigger]
	description InvocationInfo
}

// NewSyncExitActionBehaviour creates a new synchronous exit action.
func NewSyncExitActionBehaviour[TState, TTrigger comparable](
	action TransitionAction[TState, TTrigger],
	description InvocationInfo,
) *SyncExitActionBehaviour[TState, TTrigger] {
	return &SyncExitActionBehaviour[TState, TTrigger]{
		action:      action,
		description: description,
	}
}

func (s *SyncExitActionBehaviour[TState, TTrigger]) Execute(ctx context.Context, t Transition[TState, TTrigger]) error {
	if s.action != nil {
		return s.action(ctx, t)
	}
	return nil
}

func (s *SyncExitActionBehaviour[TState, TTrigger]) GetDescription() InvocationInfo {
	return s.description
}

// LifecycleActionBehaviour represents an activation or deactivation action for a state.
type LifecycleActionBehaviour[TState comparable] interface {
	// Execute executes the lifecycle action.
	Execute(ctx context.Context) error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
}

// ActivateActionBehaviour represents an activation action for a state.
type ActivateActionBehaviour[TState comparable] = LifecycleActionBehaviour[TState]

// SyncActivateActionBehaviour is a synchronous activation action.
type SyncActivateActionBehaviour[TState comparable] struct {
	action      func(ctx context.Context) error
	description InvocationInfo
}

// NewSyncActivateActionBehaviour creates a new synchronous activation action.
func NewSyncActivateActionBehaviour[TState comparable](
	action func(ctx context.Context) error,
	description InvocationInfo,
) *SyncActivateActionBehaviour[TState] {
	return &SyncActivateActionBehaviour[TState]{
		action:      action,
		description: description,
	}
}

func (s *SyncActivateActionBehaviour[TState]) Execute(ctx context.Context) error {
	if s.action != nil {
		return s.action(ctx)
	}
	return nil
}

func (s *SyncActivateActionBehaviour[TState]) GetDescription() InvocationInfo {
	return s.description
}

// DeactivateActionBehaviour represents a deactivation action for a state.
type DeactivateActionBehaviour[TState comparable] = LifecycleActionBehaviour[TState]

// SyncDeactivateActionBehaviour is a synchronous deactivation action.
type SyncDeactivateActionBehaviour[TState comparable] struct {
	action      func(ctx context.Context) error
	description InvocationInfo
}

// NewSyncDeactivateActionBehaviour creates a new synchronous deactivation action.
func NewSyncDeactivateActionBehaviour[TState comparable](
	action func(ctx context.Context) error,
	description InvocationInfo,
) *SyncDeactivateActionBehaviour[TState] {
	return &SyncDeactivateActionBehaviour[TState]{
		action:      action,
		description: description,
	}
}

func (s *SyncDeactivateActionBehaviour[TState]) Execute(ctx context.Context) error {
	if s.action != nil {
		return s.action(ctx)
	}
	return nil
}

func (s *SyncDeactivateActionBehaviour[TState]) GetDescription() InvocationInfo {
	return s.description
}
