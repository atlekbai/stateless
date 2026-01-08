package stateless

import (
	"context"
	"errors"
)

// GuardFunc is a function that evaluates a guard condition.
// It takes a context and arguments of any type, and returns nil if the condition is met,
// or an error describing why the guard failed.
type GuardFunc func(ctx context.Context, args any) error

// StateSelector is a function that determines the destination state
// based on the trigger arguments.
type StateSelector[TState comparable] func(args any) TState

// GuardCondition represents a single guard condition with its method description.
type GuardCondition struct {
	// Guard is the guard function that takes args and returns nil if the condition is met,
	// or an error describing why the guard failed.
	Guard GuardFunc

	// methodDescription contains information about the guard method.
	methodDescription InvocationInfo
}

// NewGuardCondition creates a new guard condition from a guard function that takes args.
// The guard returns nil if the condition is met, or an error describing why it failed.
func NewGuardCondition(guard GuardFunc, description InvocationInfo) GuardCondition {
	return GuardCondition{
		Guard:             guard,
		methodDescription: description,
	}
}

// Description returns the description of the guard method.
func (g GuardCondition) Description() string {
	return g.methodDescription.Description()
}

// MethodDescription returns the full method description.
func (g GuardCondition) MethodDescription() InvocationInfo {
	return g.methodDescription
}

// Evaluate evaluates the guard condition and returns an error if it fails.
// Returns nil if the guard condition is met.
func (g GuardCondition) Evaluate(ctx context.Context, args any) error {
	if g.Guard == nil {
		return nil
	}
	return g.Guard(ctx, args)
}

// TransitionGuard contains a list of guard conditions that must all be met for a transition.
type TransitionGuard struct {
	Conditions []GuardCondition
}

// EmptyTransitionGuard is a transition guard with no conditions (always passes).
var EmptyTransitionGuard = TransitionGuard{Conditions: []GuardCondition{}}

// NewTransitionGuard creates a new transition guard from a guard function.
// The guard returns nil if the condition is met, or an error describing why it failed.
func NewTransitionGuard(guard GuardFunc) TransitionGuard {
	if guard == nil {
		return EmptyTransitionGuard
	}
	return TransitionGuard{
		Conditions: []GuardCondition{
			NewGuardCondition(guard, CreateInvocationInfo(guard, "")),
		},
	}
}

// Guards returns the list of guard functions.
func (tg TransitionGuard) Guards() []GuardFunc {
	result := make([]GuardFunc, len(tg.Conditions))
	for i, c := range tg.Conditions {
		result[i] = c.Guard
	}
	return result
}

// GuardConditionsMet evaluates all guard conditions and returns an error if any fail.
// Returns nil if all guard conditions are met.
// If multiple conditions fail, returns all errors joined together.
func (tg TransitionGuard) GuardConditionsMet(ctx context.Context, args any) error {
	var errs []error
	for _, c := range tg.Conditions {
		if err := c.Evaluate(ctx, args); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// IsEmpty returns true if the transition guard has no conditions.
func (tg TransitionGuard) IsEmpty() bool {
	return len(tg.Conditions) == 0
}
