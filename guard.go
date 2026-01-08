package stateless

import "context"

// GuardFunc is a function that evaluates a guard condition.
// It takes a context and arguments of any type, and returns nil if the condition is met,
// or an error describing why the guard failed.
type GuardFunc func(ctx context.Context, args any) error

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

// IsMet returns true if the guard condition is met (returns nil).
func (g GuardCondition) IsMet(ctx context.Context, args any) bool {
	if g.Guard == nil {
		return true
	}
	return g.Guard(ctx, args) == nil
}

// Error returns the error from the guard, or nil if the guard passes.
func (g GuardCondition) Error(ctx context.Context, args any) error {
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

// GuardConditionsMet returns true if all guard conditions are met.
func (tg TransitionGuard) GuardConditionsMet(ctx context.Context, args any) bool {
	for _, c := range tg.Conditions {
		if !c.IsMet(ctx, args) {
			return false
		}
	}
	return true
}

// UnmetGuardConditions returns a list of descriptions for all guard conditions that are not met.
func (tg TransitionGuard) UnmetGuardConditions(ctx context.Context, args any) []string {
	var unmet []string
	for _, c := range tg.Conditions {
		if err := c.Error(ctx, args); err != nil {
			// Use the error message as the description
			unmet = append(unmet, err.Error())
		}
	}
	return unmet
}

// IsEmpty returns true if the transition guard has no conditions.
func (tg TransitionGuard) IsEmpty() bool {
	return len(tg.Conditions) == 0
}

// TypedGuard converts a typed guard function to a guard that takes any args.
func TypedGuard[TArgs any](guard func(context.Context, TArgs) error) GuardFunc {
	return func(ctx context.Context, args any) error {
		if args == nil {
			var zero TArgs
			return guard(ctx, zero)
		}
		typedArgs, ok := args.(TArgs)
		if !ok {
			var zero TArgs
			return guard(ctx, zero)
		}
		return guard(ctx, typedArgs)
	}
}
