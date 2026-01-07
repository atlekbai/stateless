# Plan: Unified OnEntry with Typed StateConfiguration

## Goal

Simplify the API by having ONE `OnEntry` method that always receives a typed `Transition[S, T, A]`. Each `StateConfiguration` carries its own Args type.

## Current API (Multiple Variants)

```go
sm := NewStateMachine[State, Trigger](initial)
config := sm.Configure(StateA)

// Too many options:
config.OnEntry(func(ctx context.Context) error { ... })
config.OnEntryFrom(trigger, func(ctx context.Context) error { ... })
stateless.OnEntryWithTransition[S, T, Args](config, func(ctx, t) error { ... })
stateless.OnEntryFromWithTransition[S, T, Args](config, trigger, func(ctx, t) error { ... })
```

## Proposed API (Unified)

```go
sm := NewStateMachine[State, Trigger](initial)

// Chaining works - all methods return *StateConfiguration[S, T, A]
stateless.Configure[State, Trigger, AssignArgs](sm, Assigned).
    Permit(StartWork, InProgress).
    PermitReentry(Assign).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger, AssignArgs]) error {
        fmt.Println("Assigned to:", t.Args.Assignee)
        return nil
    })

// For states without args, use NoArgs
stateless.Configure[State, Trigger, NoArgs](sm, Open).
    Permit(Assign, Assigned).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger, NoArgs]) error {
        fmt.Println("Entered from:", t.Source)
        return nil
    })
```

## Changes Required

### 1. Modify StateConfiguration struct

```go
// Before
type StateConfiguration[TState, TTrigger comparable] struct { ... }

// After
type StateConfiguration[TState, TTrigger comparable, TArgs any] struct { ... }
```

### 2. Replace sm.Configure() with standalone function

```go
// Remove method
// func (sm *StateMachine[S, T]) Configure(state S) *StateConfiguration[S, T]

// Add standalone function (required because Go doesn't allow method type params)
func Configure[TState, TTrigger comparable, TArgs any](
    sm *StateMachine[TState, TTrigger],
    state TState,
) *StateConfiguration[TState, TTrigger, TArgs]
```

### 3. Simplify OnEntry to single method

```go
// Single unified OnEntry - returns same type for chaining
func (sc *StateConfiguration[S, T, A]) OnEntry(
    action func(ctx context.Context, t Transition[S, T, A]) error,
) *StateConfiguration[S, T, A]
```

### 4. All methods return same type (chaining)

```go
func (sc *StateConfiguration[S, T, A]) Permit(trigger T, dest S) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) PermitIf(...) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) PermitReentry(trigger T) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) Ignore(trigger T) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) SubstateOf(parent S) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) OnEntry(...) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) OnExit(...) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) OnActivate(...) *StateConfiguration[S, T, A]
func (sc *StateConfiguration[S, T, A]) OnDeactivate(...) *StateConfiguration[S, T, A]
// etc.
```

### 5. Remove redundant methods/functions

Remove:
- `sm.Configure()` method - replaced by standalone `Configure[S,T,A]()`
- `OnEntryFrom()` - use `if t.Trigger == X` inside OnEntry
- `OnEntryWithTransition()` standalone - now a method
- `OnEntryFromWithTransition()` standalone - removed
- `OnExitWithTransition()` standalone - now a method

### 6. Apply same pattern to OnExit

```go
func (sc *StateConfiguration[S, T, A]) OnExit(
    action func(ctx context.Context, t Transition[S, T, A]) error,
) *StateConfiguration[S, T, A]
```

### 7. Update InternalTransition

```go
func (sc *StateConfiguration[S, T, A]) InternalTransition(
    trigger T,
    action func(ctx context.Context, t Transition[S, T, A]) error,
) *StateConfiguration[S, T, A]
```

### 8. Update OnActivate/OnDeactivate

These don't have transitions, keep simple signature:
```go
func (sc *StateConfiguration[S, T, A]) OnActivate(
    action func(ctx context.Context) error,
) *StateConfiguration[S, T, A]

func (sc *StateConfiguration[S, T, A]) OnDeactivate(
    action func(ctx context.Context) error,
) *StateConfiguration[S, T, A]
```

## Migration Example

### Before
```go
sm := NewStateMachine[State, Trigger](Open)

sm.Configure(Open).
    Permit(Assign, Assigned).
    OnEntry(func(ctx context.Context) error {
        fmt.Println("Opened")
        return nil
    })

stateless.OnEntryWithTransition[State, Trigger, AssignArgs](
    sm.Configure(Assigned),
    func(ctx context.Context, t Transition[State, Trigger, AssignArgs]) error {
        fmt.Println("Assigned to:", t.Args.Assignee)
        return nil
    })
```

### After
```go
sm := NewStateMachine[State, Trigger](Open)

stateless.Configure[State, Trigger, NoArgs](sm, Open).
    Permit(Assign, Assigned).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger, NoArgs]) error {
        fmt.Println("Opened")
        return nil
    })

stateless.Configure[State, Trigger, AssignArgs](sm, Assigned).
    Permit(StartWork, InProgress).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger, AssignArgs]) error {
        fmt.Println("Assigned to:", t.Args.Assignee)
        return nil
    })
```

## Files to Modify

1. `state_configuration.go` - Add TArgs type parameter, update all methods
2. `state_machine.go` - Remove Configure method, add standalone Configure function
3. `transition.go` - Ensure Transition[S, T, A] works correctly
4. `state_representation.go` - Update action storage/execution
5. `action_behaviour.go` - Update entry/exit action types
6. `state_machine_test.go` - Update all tests
7. `graph/` - Update graph generation if needed
8. `examples/` - Update examples
9. `README.md` - Update documentation

## Decisions Made

1. **No backwards compatibility** - Remove `sm.Configure()` entirely
2. **Chaining preserved** - All methods return `*StateConfiguration[S, T, A]`
3. **OnEntryFrom removed** - Use conditional `if t.Trigger == X` in OnEntry

## Risks

- Breaking change for all existing users
- More verbose for simple cases (must use `NoArgs`)
- Every state configuration requires explicit type parameters
