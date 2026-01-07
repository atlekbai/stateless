# Plan: Unified Actions with Transition (Args as any)

## Goal

Simplify the API so all Fire()-triggered actions receive `Transition[S, T]` with `Args` as `any`. User does type assertion when needed.

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

sm.Configure(Assigned).
    Permit(StartWork, InProgress).
    PermitReentry(Assign).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger]) error {
        // Type assertion when needed
        if args, ok := t.Args.(AssignArgs); ok {
            fmt.Println("Assigned to:", args.Assignee)
        }
        return nil
    }).
    OnExit(func(ctx context.Context, t Transition[State, Trigger]) error {
        fmt.Println("Leaving Assigned, going to:", t.Destination)
        return nil
    }).
    InternalTransition(LogEvent, func(ctx context.Context, t Transition[State, Trigger]) error {
        if args, ok := t.Args.(LogArgs); ok {
            fmt.Println("Log:", args.Message)
        }
        return nil
    })
```

## Changes Required

### 1. Keep StateConfiguration as is (no TArgs)

```go
// No change needed
type StateConfiguration[TState, TTrigger comparable] struct { ... }
```

### 2. Keep sm.Configure() method

```go
// No change needed
func (sm *StateMachine[S, T]) Configure(state S) *StateConfiguration[S, T]
```

### 3. Transition with Args as any

```go
type Transition[TState, TTrigger comparable] struct {
    Source      TState
    Destination TState
    Trigger     TTrigger
    Args        any  // User does type assertion
}
```

### 4. OnEntry receives Transition

```go
func (sc *StateConfiguration[S, T]) OnEntry(
    action func(ctx context.Context, t Transition[S, T]) error,
) *StateConfiguration[S, T]
```

### 5. OnExit receives Transition

```go
func (sc *StateConfiguration[S, T]) OnExit(
    action func(ctx context.Context, t Transition[S, T]) error,
) *StateConfiguration[S, T]
```

### 6. InternalTransition receives Transition

```go
func (sc *StateConfiguration[S, T]) InternalTransition(
    trigger T,
    action func(ctx context.Context, t Transition[S, T]) error,
) *StateConfiguration[S, T]
```

### 7. Remove redundant methods/functions

Remove:
- `OnEntryFrom()` - use `if t.Trigger == X` inside OnEntry
- `OnEntryWithTransition()` standalone function
- `OnEntryFromWithTransition()` standalone function
- `OnExitWithTransition()` standalone function

### 8. OnActivate/OnDeactivate (no Transition - not triggered by Fire)

```go
func (sc *StateConfiguration[S, T]) OnActivate(
    action func(ctx context.Context) error,
) *StateConfiguration[S, T]

func (sc *StateConfiguration[S, T]) OnDeactivate(
    action func(ctx context.Context) error,
) *StateConfiguration[S, T]
```

## Summary of Action Signatures

| Method | Signature | Triggered by |
|--------|-----------|--------------|
| `OnEntry` | `func(ctx, t Transition[S,T]) error` | Fire() |
| `OnExit` | `func(ctx, t Transition[S,T]) error` | Fire() |
| `InternalTransition` | `func(ctx, t Transition[S,T]) error` | Fire() |
| `OnActivate` | `func(ctx) error` | Activate() |
| `OnDeactivate` | `func(ctx) error` | Deactivate() |

**Consistent rule**: All actions triggered by `Fire()` receive `Transition`. Actions triggered by other methods (`Activate`/`Deactivate`) don't.

## Migration Example

### Before
```go
sm := NewStateMachine[State, Trigger](Open)

sm.Configure(Open).
    Permit(Assign, Assigned).
    OnEntry(func(ctx context.Context) error {
        fmt.Println("Opened")
        return nil
    }).
    OnExit(func(ctx context.Context) error {
        fmt.Println("Leaving Open")
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

sm.Configure(Open).
    Permit(Assign, Assigned).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger]) error {
        fmt.Println("Opened, from:", t.Source)
        return nil
    }).
    OnExit(func(ctx context.Context, t Transition[State, Trigger]) error {
        fmt.Println("Leaving Open, trigger:", t.Trigger)
        return nil
    })

sm.Configure(Assigned).
    Permit(StartWork, InProgress).
    OnEntry(func(ctx context.Context, t Transition[State, Trigger]) error {
        if args, ok := t.Args.(AssignArgs); ok {
            fmt.Println("Assigned to:", args.Assignee)
        }
        return nil
    })
```

## Files to Modify

1. `transition.go` - Remove TArgs type parameter, Args becomes `any`
2. `state_configuration.go` - Update OnEntry/OnExit/InternalTransition signatures, remove OnEntryFrom etc.
3. `state_machine.go` - Remove standalone generic functions
4. `state_representation.go` - Update action storage/execution
5. `action_behaviour.go` - Update entry/exit action types
6. `state_machine_test.go` - Update all tests
7. `graph/` - Update if needed
8. `examples/` - Update examples
9. `README.md` - Update documentation

## Decisions Made

1. **Args as `any`** - User does type assertion when needed
2. **Keep `sm.Configure()`** - No extra type parameter needed
3. **All Fire()-triggered actions receive Transition** - OnEntry, OnExit, InternalTransition
4. **OnActivate/OnDeactivate** - Just context (not triggered by Fire)
5. **Remove OnEntryFrom** - Use `if t.Trigger == X` in OnEntry

## Benefits

- Simpler API - no complex generics
- Consistent pattern - Fire() actions get Transition
- Keeps chaining working naturally
- Type assertion is common Go pattern
- No standalone generic functions needed
