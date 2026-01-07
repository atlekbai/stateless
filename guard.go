package stateless

// GuardCondition represents a single guard condition with its method description.
type GuardCondition struct {
	// Guard is the guard function that takes args and returns true if the condition is met.
	Guard func(args any) bool

	// methodDescription contains information about the guard method.
	methodDescription InvocationInfo
}

// NewGuardCondition creates a new guard condition from a parameterless guard function.
func NewGuardCondition(guard func() bool, description InvocationInfo) GuardCondition {
	return GuardCondition{
		Guard:             func(args any) bool { return guard() },
		methodDescription: description,
	}
}

// NewGuardConditionWithArgs creates a new guard condition from a guard function that takes args.
func NewGuardConditionWithArgs(guard func(args any) bool, description InvocationInfo) GuardCondition {
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
			NewGuardCondition(guard, CreateInvocationInfo(guard, description)),
		},
	}
}

// NewTransitionGuardWithArgs creates a new transition guard from a guard function that takes args.
func NewTransitionGuardWithArgs(guard func(args any) bool, description string) TransitionGuard {
	if guard == nil {
		return EmptyTransitionGuard
	}
	return TransitionGuard{
		Conditions: []GuardCondition{
			NewGuardConditionWithArgs(guard, CreateInvocationInfo(guard, description)),
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
		conditions[i] = NewGuardCondition(g.Guard, CreateInvocationInfo(g.Guard, g.Description))
	}
	return TransitionGuard{Conditions: conditions}
}

// NewTransitionGuardFromTuplesWithArgs creates a transition guard from multiple guard/description pairs
// where guards take args.
func NewTransitionGuardFromTuplesWithArgs(guards []struct {
	Guard       func(args any) bool
	Description string
}) TransitionGuard {
	if len(guards) == 0 {
		return EmptyTransitionGuard
	}
	conditions := make([]GuardCondition, len(guards))
	for i, g := range guards {
		conditions[i] = NewGuardConditionWithArgs(g.Guard, CreateInvocationInfo(g.Guard, g.Description))
	}
	return TransitionGuard{Conditions: conditions}
}

// Guards returns the list of guard functions.
func (tg TransitionGuard) Guards() []func(args any) bool {
	result := make([]func(args any) bool, len(tg.Conditions))
	for i, c := range tg.Conditions {
		result[i] = c.Guard
	}
	return result
}

// GuardConditionsMet returns true if all guard conditions are met.
func (tg TransitionGuard) GuardConditionsMet(args any) bool {
	for _, c := range tg.Conditions {
		if c.Guard != nil && !c.Guard(args) {
			return false
		}
	}
	return true
}

// UnmetGuardConditions returns a list of descriptions for all guard conditions that are not met.
func (tg TransitionGuard) UnmetGuardConditions(args any) []string {
	var unmet []string
	for _, c := range tg.Conditions {
		if !c.Guard(args) {
			unmet = append(unmet, c.Description())
		}
	}
	return unmet
}

// IsEmpty returns true if the transition guard has no conditions.
func (tg TransitionGuard) IsEmpty() bool {
	return len(tg.Conditions) == 0
}

// TypedGuard converts a typed guard function to a guard that takes any args.
func TypedGuard[TArgs any](guard func(TArgs) bool) func(any) bool {
	return func(args any) bool {
		if args == nil {
			var zero TArgs
			return guard(zero)
		}
		typedArgs, ok := args.(TArgs)
		if !ok {
			var zero TArgs
			return guard(zero)
		}
		return guard(typedArgs)
	}
}
