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

// Configure returns typed StateConfiguration
assignedConfig := stateless.Configure[State, Trigger, AssignArgs](sm, Assigned)
openConfig := stateless.Configure[State, Trigger, NoArgs](sm, Open)

// Single OnEntry method - always receives Transition
assignedConfig.OnEntry(func(ctx context.Context, t Transition[State, Trigger, AssignArgs]) error {
    fmt.Println("Assigned to:", t.Args.Assignee)
    return nil
})

// For states without args, use NoArgs
openConfig.OnEntry(func(ctx context.Context, t Transition[State, Trigger, NoArgs]) error {
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

### 2. Add standalone Configure function

```go
// Standalone function (Go doesn't allow method type params)
func Configure[TState, TTrigger comparable, TArgs any](
    sm *StateMachine[TState, TTrigger],
    state TState,
) *StateConfiguration[TState, TTrigger, TArgs]
```

### 3. Simplify OnEntry to single method

```go
// Single unified OnEntry
func (sc *StateConfiguration[S, T, A]) OnEntry(
    action func(ctx context.Context, t Transition[S, T, A]) error,
) *StateConfiguration[S, T, A]
```

### 4. Remove redundant methods

Remove:
- `OnEntryFrom` - use `if t.Trigger == X` in OnEntry
- `OnEntryWithTransition` (standalone) - now a method
- `OnEntryFromWithTransition` (standalone) - use conditional in OnEntry

### 5. Apply same pattern to OnExit

```go
func (sc *StateConfiguration[S, T, A]) OnExit(
    action func(ctx context.Context, t Transition[S, T, A]) error,
) *StateConfiguration[S, T, A]
```

Remove:
- `OnExitWithTransition` (standalone)

### 6. Update internal action storage

Entry/exit actions stored with typed signature matching the config's Args type.

### 7. Keep sm.Configure() for backwards compatibility (optional)

```go
// Returns StateConfiguration[S, T, NoArgs] for simple cases
func (sm *StateMachine[S, T]) Configure(state S) *StateConfiguration[S, T, NoArgs]
```

## Migration Example

### Before
```go
sm := NewStateMachine[State, Trigger](Open)

sm.Configure(Open).
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
    OnEntry(func(ctx context.Context, t Transition[State, Trigger, NoArgs]) error {
        fmt.Println("Opened")
        return nil
    })

stateless.Configure[State, Trigger, AssignArgs](sm, Assigned).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger, AssignArgs]) error {
        fmt.Println("Assigned to:", t.Args.Assignee)
        return nil
    })
```

## Files to Modify

1. `state_configuration.go` - Add TArgs type parameter, simplify methods
2. `state_machine.go` - Add standalone Configure function
3. `transition.go` - Ensure Transition[S, T, A] works correctly
4. `state_representation.go` - Update action storage/execution
5. `action_behaviour.go` - Update entry/exit action types
6. `state_machine_test.go` - Update all tests
7. `graph/` - Update graph generation if needed
8. `examples/` - Update examples
9. `README.md` - Update documentation

## Open Questions

1. **Backwards compatibility**: Keep `sm.Configure()` returning `StateConfiguration[S, T, NoArgs]`?
2. **OnEntryFrom behavior**: Remove entirely or keep as convenience that filters by trigger?
3. **Chaining**: How to handle `Configure().Permit().OnEntry()` chaining with different return types?

## Risks

- Breaking change for all existing users
- More verbose for simple cases (must use `NoArgs`)
- Chaining might become awkward if Permit returns different type than OnEntry needs
