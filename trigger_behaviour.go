package stateless

import "context"

// TriggerBehaviour is the base interface for all trigger behaviours.
type TriggerBehaviour[TState, TTrigger comparable] interface {
	// GetTrigger returns the trigger associated with this behaviour.
	GetTrigger() TTrigger

	// GetGuard returns the transition guard for this trigger.
	GetGuard() TransitionGuard

	// GuardConditionsMet returns true if all guard conditions are met.
	GuardConditionsMet(ctx context.Context, args any) bool

	// UnmetGuardConditions returns the descriptions of all unmet guard conditions.
	UnmetGuardConditions(ctx context.Context, args any) []string
}

// triggerBehaviourBase provides the base implementation for trigger behaviours.
type triggerBehaviourBase[TState, TTrigger comparable] struct {
	trigger TTrigger
	guard   TransitionGuard
}

func (t *triggerBehaviourBase[TState, TTrigger]) GetTrigger() TTrigger {
	return t.trigger
}

func (t *triggerBehaviourBase[TState, TTrigger]) GetGuard() TransitionGuard {
	return t.guard
}

func (t *triggerBehaviourBase[TState, TTrigger]) GuardConditionsMet(ctx context.Context, args any) bool {
	return t.guard.GuardConditionsMet(ctx, args)
}

func (t *triggerBehaviourBase[TState, TTrigger]) UnmetGuardConditions(ctx context.Context, args any) []string {
	return t.guard.UnmetGuardConditions(ctx, args)
}

// TransitioningTriggerBehaviour represents a transition to a fixed destination state.
type TransitioningTriggerBehaviour[TState, TTrigger comparable] struct {
	triggerBehaviourBase[TState, TTrigger]

	Destination TState
}

// NewTransitioningTriggerBehaviour creates a new transitioning trigger behaviour.
func NewTransitioningTriggerBehaviour[TState, TTrigger comparable](
	trigger TTrigger,
	destination TState,
	guard TransitionGuard,
) *TransitioningTriggerBehaviour[TState, TTrigger] {
	return &TransitioningTriggerBehaviour[TState, TTrigger]{
		triggerBehaviourBase: triggerBehaviourBase[TState, TTrigger]{
			trigger: trigger,
			guard:   guard,
		},
		Destination: destination,
	}
}

// ReentryTriggerBehaviour represents a reentry transition (state exits and re-enters itself).
type ReentryTriggerBehaviour[TState, TTrigger comparable] struct {
	triggerBehaviourBase[TState, TTrigger]

	Destination TState
}

// NewReentryTriggerBehaviour creates a new reentry trigger behaviour.
func NewReentryTriggerBehaviour[TState, TTrigger comparable](
	trigger TTrigger,
	destination TState,
	guard TransitionGuard,
) *ReentryTriggerBehaviour[TState, TTrigger] {
	return &ReentryTriggerBehaviour[TState, TTrigger]{
		triggerBehaviourBase: triggerBehaviourBase[TState, TTrigger]{
			trigger: trigger,
			guard:   guard,
		},
		Destination: destination,
	}
}

// IgnoredTriggerBehaviour represents a trigger that should be ignored.
type IgnoredTriggerBehaviour[TState, TTrigger comparable] struct {
	triggerBehaviourBase[TState, TTrigger]
}

// NewIgnoredTriggerBehaviour creates a new ignored trigger behaviour.
func NewIgnoredTriggerBehaviour[TState, TTrigger comparable](
	trigger TTrigger,
	guard TransitionGuard,
) *IgnoredTriggerBehaviour[TState, TTrigger] {
	return &IgnoredTriggerBehaviour[TState, TTrigger]{
		triggerBehaviourBase: triggerBehaviourBase[TState, TTrigger]{
			trigger: trigger,
			guard:   guard,
		},
	}
}

// DynamicTriggerBehaviour represents a transition to a dynamically determined state.
type DynamicTriggerBehaviour[TState, TTrigger comparable] struct {
	triggerBehaviourBase[TState, TTrigger]

	destination    func(args any) TState
	TransitionInfo DynamicTransitionInfo
}

// NewDynamicTriggerBehaviour creates a new dynamic trigger behaviour.
func NewDynamicTriggerBehaviour[TState, TTrigger comparable](
	trigger TTrigger,
	destination func(args any) TState,
	guard TransitionGuard,
	info DynamicTransitionInfo,
) *DynamicTriggerBehaviour[TState, TTrigger] {
	return &DynamicTriggerBehaviour[TState, TTrigger]{
		triggerBehaviourBase: triggerBehaviourBase[TState, TTrigger]{
			trigger: trigger,
			guard:   guard,
		},
		destination:    destination,
		TransitionInfo: info,
	}
}

// GetDestinationState returns the destination state based on the given arguments.
func (d *DynamicTriggerBehaviour[TState, TTrigger]) GetDestinationState(args any) TState {
	return d.destination(args)
}

// InternalTriggerBehaviour represents an internal transition that doesn't exit/enter the state.
type InternalTriggerBehaviour[TState, TTrigger comparable] interface {
	TriggerBehaviour[TState, TTrigger]
	Execute(ctx context.Context, transition Transition[TState, TTrigger]) error
}

// SyncInternalTriggerBehaviour represents a synchronous internal transition.
type SyncInternalTriggerBehaviour[TState, TTrigger comparable] struct {
	triggerBehaviourBase[TState, TTrigger]

	internalAction TransitionAction[TState, TTrigger]
}

// NewSyncInternalTriggerBehaviour creates a new synchronous internal trigger behaviour.
func NewSyncInternalTriggerBehaviour[TState, TTrigger comparable](
	trigger TTrigger,
	guard TransitionGuard,
	internalAction TransitionAction[TState, TTrigger],
) *SyncInternalTriggerBehaviour[TState, TTrigger] {
	return &SyncInternalTriggerBehaviour[TState, TTrigger]{
		triggerBehaviourBase: triggerBehaviourBase[TState, TTrigger]{
			trigger: trigger,
			guard:   guard,
		},
		internalAction: internalAction,
	}
}

// Execute executes the internal action.
func (s *SyncInternalTriggerBehaviour[TState, TTrigger]) Execute(
	ctx context.Context,
	transition Transition[TState, TTrigger],
) error {
	if s.internalAction != nil {
		return s.internalAction(ctx, transition)
	}
	return nil
}

// TriggerBehaviourResult represents the result of finding a trigger behaviour.
type TriggerBehaviourResult[TState, TTrigger comparable] struct {
	// Handler is the trigger behaviour that was found.
	Handler TriggerBehaviour[TState, TTrigger]

	// UnmetGuardConditions contains descriptions of any unmet guard conditions.
	UnmetGuardConditions []string

	// MultipleHandlersFound indicates if multiple handlers matched (configuration error).
	MultipleHandlersFound bool
}
