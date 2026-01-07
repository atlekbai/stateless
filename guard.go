package stateless

// GuardCondition represents a single guard condition with its method description.
type GuardCondition struct {
	// Guard is the guard function that takes parameters and returns true if the condition is met.
	Guard func(args ...any) bool

	// methodDescription contains information about the guard method.
	methodDescription InvocationInfo
}

// NewGuardCondition creates a new guard condition from a parameterless guard function.
func NewGuardCondition(guard func() bool, description InvocationInfo) GuardCondition {
	return GuardCondition{
		Guard:             func(args ...any) bool { return guard() },
		methodDescription: description,
	}
}

// NewGuardConditionWithArgs creates a new guard condition from a guard function that takes parameters.
func NewGuardConditionWithArgs(guard func(args ...any) bool, description InvocationInfo) GuardCondition {
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

// TransitionGuard contains a list of guard conditions that must all be met for a transition.
type TransitionGuard struct {
	Conditions []GuardCondition
}

// EmptyTransitionGuard is a transition guard with no conditions (always passes).
var EmptyTransitionGuard = TransitionGuard{Conditions: []GuardCondition{}}

// NewTransitionGuard creates a new transition guard from a parameterless guard function.
func NewTransitionGuard(guard func() bool, description string) TransitionGuard {
	if guard == nil {
		return EmptyTransitionGuard
	}
	return TransitionGuard{
		Conditions: []GuardCondition{
			NewGuardCondition(guard, CreateInvocationInfo(guard, description, TimingSynchronous)),
		},
	}
}

// NewTransitionGuardWithArgs creates a new transition guard from a guard function that takes parameters.
func NewTransitionGuardWithArgs(guard func(args ...any) bool, description string) TransitionGuard {
	if guard == nil {
		return EmptyTransitionGuard
	}
	return TransitionGuard{
		Conditions: []GuardCondition{
			NewGuardConditionWithArgs(guard, CreateInvocationInfo(guard, description, TimingSynchronous)),
		},
	}
}

// NewTransitionGuardFromTuples creates a transition guard from multiple guard/description pairs.
func NewTransitionGuardFromTuples(guards []struct {
	Guard       func() bool
	Description string
}) TransitionGuard {
	if len(guards) == 0 {
		return EmptyTransitionGuard
	}
	conditions := make([]GuardCondition, len(guards))
	for i, g := range guards {
		conditions[i] = NewGuardCondition(g.Guard, CreateInvocationInfo(g.Guard, g.Description, TimingSynchronous))
	}
	return TransitionGuard{Conditions: conditions}
}

// NewTransitionGuardFromTuplesWithArgs creates a transition guard from multiple guard/description pairs
// where guards take parameters.
func NewTransitionGuardFromTuplesWithArgs(guards []struct {
	Guard       func(args ...any) bool
	Description string
}) TransitionGuard {
	if len(guards) == 0 {
		return EmptyTransitionGuard
	}
	conditions := make([]GuardCondition, len(guards))
	for i, g := range guards {
		conditions[i] = NewGuardConditionWithArgs(g.Guard, CreateInvocationInfo(g.Guard, g.Description, TimingSynchronous))
	}
	return TransitionGuard{Conditions: conditions}
}

// Guards returns the list of guard functions.
func (tg TransitionGuard) Guards() []func(args ...any) bool {
	result := make([]func(args ...any) bool, len(tg.Conditions))
	for i, c := range tg.Conditions {
		result[i] = c.Guard
	}
	return result
}

// GuardConditionsMet returns true if all guard conditions are met.
func (tg TransitionGuard) GuardConditionsMet(args ...any) bool {
	for _, c := range tg.Conditions {
		if c.Guard != nil && !c.Guard(args...) {
			return false
		}
	}
	return true
}

// UnmetGuardConditions returns a list of descriptions for all guard conditions that are not met.
func (tg TransitionGuard) UnmetGuardConditions(args ...any) []string {
	var unmet []string
	for _, c := range tg.Conditions {
		if !c.Guard(args...) {
			unmet = append(unmet, c.Description())
		}
	}
	return unmet
}

// IsEmpty returns true if the transition guard has no conditions.
func (tg TransitionGuard) IsEmpty() bool {
	return len(tg.Conditions) == 0
}

// ToPackedGuard converts a typed guard function to a packed guard function.
func ToPackedGuard[TArg0 any](guard func(TArg0) bool) func(args ...any) bool {
	return func(args ...any) bool {
		if len(args) == 0 {
			var zero TArg0
			return guard(zero)
		}
		return guard(args[0].(TArg0))
	}
}

// ToPackedGuard2 converts a typed guard function with two arguments to a packed guard function.
func ToPackedGuard2[TArg0, TArg1 any](guard func(TArg0, TArg1) bool) func(args ...any) bool {
	return func(args ...any) bool {
		var arg0 TArg0
		var arg1 TArg1
		if len(args) > 0 {
			arg0 = args[0].(TArg0)
		}
		if len(args) > 1 {
			arg1 = args[1].(TArg1)
		}
		return guard(arg0, arg1)
	}
}

// ToPackedGuard3 converts a typed guard function with three arguments to a packed guard function.
func ToPackedGuard3[TArg0, TArg1, TArg2 any](guard func(TArg0, TArg1, TArg2) bool) func(args ...any) bool {
	return func(args ...any) bool {
		var arg0 TArg0
		var arg1 TArg1
		var arg2 TArg2
		if len(args) > 0 {
			arg0 = args[0].(TArg0)
		}
		if len(args) > 1 {
			arg1 = args[1].(TArg1)
		}
		if len(args) > 2 {
			arg2 = args[2].(TArg2)
		}
		return guard(arg0, arg1, arg2)
	}
}
