package stateless

// EntryActionBehaviour represents an entry action for a state.
type EntryActionBehaviour[TState, TTrigger comparable] interface {
	// Execute executes the entry action.
	Execute(transition internalTransition[TState, TTrigger]) error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
	// GetFromTrigger returns the trigger this action is bound to (nil if not bound).
	GetFromTrigger() *TTrigger
}

// SyncEntryActionBehaviour is a synchronous entry action.
type SyncEntryActionBehaviour[TState, TTrigger comparable] struct {
	action      func(transition internalTransition[TState, TTrigger])
	description InvocationInfo
}

// NewSyncEntryActionBehaviour creates a new synchronous entry action.
func NewSyncEntryActionBehaviour[TState, TTrigger comparable](
	action func(transition internalTransition[TState, TTrigger]),
	description InvocationInfo,
) *SyncEntryActionBehaviour[TState, TTrigger] {
	return &SyncEntryActionBehaviour[TState, TTrigger]{
		action:      action,
		description: description,
	}
}

func (s *SyncEntryActionBehaviour[TState, TTrigger]) Execute(transition internalTransition[TState, TTrigger]) error {
	if s.action != nil {
		s.action(transition)
	}
	return nil
}

func (s *SyncEntryActionBehaviour[TState, TTrigger]) GetDescription() InvocationInfo {
	return s.description
}

func (s *SyncEntryActionBehaviour[TState, TTrigger]) GetFromTrigger() *TTrigger {
	return nil
}

// SyncEntryActionBehaviourFrom is a synchronous entry action that only executes for a specific trigger.
type SyncEntryActionBehaviourFrom[TState, TTrigger comparable] struct {
	*SyncEntryActionBehaviour[TState, TTrigger]
	trigger TTrigger
}

// NewSyncEntryActionBehaviourFrom creates a new synchronous entry action bound to a specific trigger.
func NewSyncEntryActionBehaviourFrom[TState, TTrigger comparable](
	trigger TTrigger,
	action func(transition internalTransition[TState, TTrigger]),
	description InvocationInfo,
) *SyncEntryActionBehaviourFrom[TState, TTrigger] {
	return &SyncEntryActionBehaviourFrom[TState, TTrigger]{
		SyncEntryActionBehaviour: NewSyncEntryActionBehaviour(action, description),
		trigger:                  trigger,
	}
}

func (s *SyncEntryActionBehaviourFrom[TState, TTrigger]) Execute(transition internalTransition[TState, TTrigger]) error {
	if transition.Trigger == s.trigger {
		return s.SyncEntryActionBehaviour.Execute(transition)
	}
	return nil
}

func (s *SyncEntryActionBehaviourFrom[TState, TTrigger]) GetFromTrigger() *TTrigger {
	return &s.trigger
}

// ExitActionBehaviour represents an exit action for a state.
type ExitActionBehaviour[TState, TTrigger comparable] interface {
	// Execute executes the exit action.
	Execute(transition internalTransition[TState, TTrigger]) error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
}

// SyncExitActionBehaviour is a synchronous exit action.
type SyncExitActionBehaviour[TState, TTrigger comparable] struct {
	action      func(transition internalTransition[TState, TTrigger])
	description InvocationInfo
}

// NewSyncExitActionBehaviour creates a new synchronous exit action.
func NewSyncExitActionBehaviour[TState, TTrigger comparable](
	action func(transition internalTransition[TState, TTrigger]),
	description InvocationInfo,
) *SyncExitActionBehaviour[TState, TTrigger] {
	return &SyncExitActionBehaviour[TState, TTrigger]{
		action:      action,
		description: description,
	}
}

func (s *SyncExitActionBehaviour[TState, TTrigger]) Execute(transition internalTransition[TState, TTrigger]) error {
	if s.action != nil {
		s.action(transition)
	}
	return nil
}

func (s *SyncExitActionBehaviour[TState, TTrigger]) GetDescription() InvocationInfo {
	return s.description
}

// ActivateActionBehaviour represents an activation action for a state.
type ActivateActionBehaviour[TState comparable] interface {
	// Execute executes the activation action.
	Execute() error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
}

// SyncActivateActionBehaviour is a synchronous activation action.
type SyncActivateActionBehaviour[TState comparable] struct {
	action      func()
	description InvocationInfo
}

// NewSyncActivateActionBehaviour creates a new synchronous activation action.
func NewSyncActivateActionBehaviour[TState comparable](
	action func(),
	description InvocationInfo,
) *SyncActivateActionBehaviour[TState] {
	return &SyncActivateActionBehaviour[TState]{
		action:      action,
		description: description,
	}
}

func (s *SyncActivateActionBehaviour[TState]) Execute() error {
	if s.action != nil {
		s.action()
	}
	return nil
}

func (s *SyncActivateActionBehaviour[TState]) GetDescription() InvocationInfo {
	return s.description
}

// DeactivateActionBehaviour represents a deactivation action for a state.
type DeactivateActionBehaviour[TState comparable] interface {
	// Execute executes the deactivation action.
	Execute() error
	// GetDescription returns the description of the action.
	GetDescription() InvocationInfo
}

// SyncDeactivateActionBehaviour is a synchronous deactivation action.
type SyncDeactivateActionBehaviour[TState comparable] struct {
	action      func()
	description InvocationInfo
}

// NewSyncDeactivateActionBehaviour creates a new synchronous deactivation action.
func NewSyncDeactivateActionBehaviour[TState comparable](
	action func(),
	description InvocationInfo,
) *SyncDeactivateActionBehaviour[TState] {
	return &SyncDeactivateActionBehaviour[TState]{
		action:      action,
		description: description,
	}
}

func (s *SyncDeactivateActionBehaviour[TState]) Execute() error {
	if s.action != nil {
		s.action()
	}
	return nil
}

func (s *SyncDeactivateActionBehaviour[TState]) GetDescription() InvocationInfo {
	return s.description
}
