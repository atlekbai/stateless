# Stateless - A State Machine Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/atlekbai/stateless.svg)](https://pkg.go.dev/github.com/atlekbai/stateless)

A feature-complete, generic state machine library for Go, inspired by the [.NET Stateless library](https://github.com/dotnet-state-machine/stateless).

## Features

- **Generic Types**: Use any comparable type for states and triggers
- **Fluent Configuration API**: Easy-to-read state machine configuration
- **Guard Conditions**: Conditional transitions with guard functions
- **Entry/Exit Actions**: Execute actions when entering or exiting states
- **Activation/Deactivation**: Lifecycle hooks for state machine activation
- **Hierarchical States**: Support for substates and superstates
- **Parameterized Triggers**: Pass data with trigger firing
- **Dynamic Transitions**: Determine destination state at runtime
- **Reentry Transitions**: Support for self-transitions with action execution
- **Internal Transitions**: Actions without state change
- **Firing Modes**: Immediate or queued trigger processing
- **Introspection**: Reflect on state machine configuration
- **Graph Generation**: Export to DOT (Graphviz) and Mermaid formats
- **Thread-Safe**: Safe for concurrent access with queued firing mode

## Installation

```bash
go get github.com/atlekbai/stateless
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/atlekbai/stateless"
)

type State int
type Trigger int

const (
    Off State = iota
    On
)

const (
    Toggle Trigger = iota
)

func main() {
    // Create a new state machine
    sm := stateless.NewStateMachine[State, Trigger](Off)

    // Configure states
    sm.Configure(Off).
        Permit(Toggle, On).
        OnExit(func(ctx context.Context) error {
            fmt.Println("Light turning on...")
            return nil
        })

    sm.Configure(On).
        Permit(Toggle, Off).
        OnEntry(func(ctx context.Context) error {
            fmt.Println("Light is on!")
            return nil
        })

    // Fire triggers
    sm.Fire(Toggle, nil) // Off -> On
    sm.Fire(Toggle, nil) // On -> Off
}
```

## Configuration

### Basic Transitions

```go
sm.Configure(StateA).
    Permit(TriggerX, StateB)  // TriggerX causes transition to StateB
```

### Conditional Transitions (Guards)

```go
sm.Configure(StateA).
    PermitIf(TriggerX, StateB, func() bool {
        return someCondition
    }, "Guard description")
```

### Ignored Triggers

```go
sm.Configure(StateA).
    Ignore(TriggerX)  // TriggerX does nothing in StateA
```

### Reentry Transitions

```go
sm.Configure(StateA).
    PermitReentry(TriggerX)  // TriggerX causes exit and entry of StateA
```

### Internal Transitions

```go
sm.Configure(StateA).
    InternalTransition(TriggerX, func(ctx context.Context) error {
        // Action executed without state change
        return nil
    })
```

### Dynamic Transitions

```go
sm.Configure(StateA).
    PermitDynamic(TriggerX, func() State {
        if someCondition {
            return StateB
        }
        return StateC
    })
```

## Entry and Exit Actions

```go
sm.Configure(StateA).
    OnEntry(func(ctx context.Context) error {
        fmt.Println("Entering StateA")
        return nil
    }).
    OnExit(func(ctx context.Context) error {
        fmt.Println("Exiting StateA")
        return nil
    }).
    OnEntryFrom(TriggerX, func(ctx context.Context) error {
        fmt.Println("Entered from TriggerX")
        return nil
    })

// For typed entry actions with transition info, use generic functions:
stateless.OnEntryWithTransition[State, Trigger, stateless.NoArgs](
    sm.Configure(StateA),
    func(ctx context.Context, t stateless.Transition[State, Trigger, stateless.NoArgs]) error {
        fmt.Printf("Entered from %v\n", t.Source)
        return nil
    })
```

## Hierarchical States

```go
// StateC is a substate of StateB
sm.Configure(StateB).
    Permit(TriggerX, StateA)

sm.Configure(StateC).
    SubstateOf(StateB)  // StateC inherits TriggerX -> StateA

// Check if in superstate
sm.IsInState(StateB)  // true when in StateB or StateC
```

## Parameterized Triggers

```go
// Define a struct for trigger arguments
type CallArgs struct {
    CallerID string
}

// Use typed entry action with transition info
stateless.OnEntryWithTransition[State, Trigger, CallArgs](
    sm.Configure(StateB),
    func(ctx context.Context, t stateless.Transition[State, Trigger, CallArgs]) error {
        fmt.Printf("Call from: %s\n", t.Args.CallerID)
        return nil
    })

// Fire with struct argument
sm.Fire(TriggerX, CallArgs{CallerID: "555-1234"})
```

## Activation and Deactivation

```go
sm.Configure(StateA).
    OnActivate(func(ctx context.Context) error {
        fmt.Println("State machine activated in StateA")
        return nil
    }).
    OnDeactivate(func(ctx context.Context) error {
        fmt.Println("State machine deactivated in StateA")
        return nil
    })

sm.Activate(context.Background())    // Calls OnActivate
sm.Deactivate(context.Background())  // Calls OnDeactivate
```

## Event Handlers

```go
// Called when a transition occurs (use generic function)
stateless.OnTransitioned[State, Trigger, stateless.NoArgs](sm,
    func(t stateless.Transition[State, Trigger, stateless.NoArgs]) {
        fmt.Printf("%v -> %v\n", t.Source, t.Destination)
    })

// Called after all transition actions complete
stateless.OnTransitionCompleted[State, Trigger, stateless.NoArgs](sm,
    func(t stateless.Transition[State, Trigger, stateless.NoArgs]) {
        fmt.Println("Transition completed")
    })

// Handle unhandled triggers
sm.OnUnhandledTrigger(func(state State, trigger Trigger, guards []string) {
    fmt.Printf("Unhandled trigger %v in state %v\n", trigger, state)
})
```

## Firing Modes

```go
// Immediate mode (default) - triggers processed synchronously
sm := stateless.NewStateMachine[State, Trigger](StateA)

// Queued mode - triggers queued and processed one at a time
sm := stateless.NewStateMachineWithMode[State, Trigger](StateA, stateless.FiringQueued)
```

## External State Storage

```go
var currentState State = StateA

sm := stateless.NewStateMachineWithExternalStorage[State, Trigger](
    func() State { return currentState },     // Accessor
    func(s State) { currentState = s },       // Mutator
)
```

## Introspection

```go
// Check current state
state := sm.State()

// Check if trigger can be fired
canFire := sm.CanFire(TriggerX)

// Get permitted triggers
triggers := sm.GetPermittedTriggers(nil)

// Check if in a state (including substates)
isInState := sm.IsInState(StateA)

// Get full state machine info
info := sm.GetInfo()
```

## Graph Generation

### DOT Graph (Graphviz)

```go
import "github.com/atlekbai/stateless/graph"

info := sm.GetInfo()
dot := graph.UmlDotGraph(info)
fmt.Println(dot)
```

### Mermaid Graph

```go
import "github.com/atlekbai/stateless/graph"

info := sm.GetInfo()
direction := graph.LeftToRight
mermaid := graph.MermaidGraph(info, &direction)
fmt.Println(mermaid)
```

## Context Support

All actions receive a `context.Context` parameter. Use `FireCtx` to pass a custom context:

```go
import "context"

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := sm.FireCtx(ctx, TriggerX, nil)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/atlekbai/stateless"
    "github.com/atlekbai/stateless/graph"
)

type PhoneState int
type PhoneTrigger int

const (
    OffHook PhoneState = iota
    Ringing
    Connected
    OnHold
)

const (
    CallDialed PhoneTrigger = iota
    CallConnected
    PlacedOnHold
    TakenOffHold
    HungUp
)

func main() {
    sm := stateless.NewStateMachine[PhoneState, PhoneTrigger](OffHook)

    sm.Configure(OffHook).
        Permit(CallDialed, Ringing)

    sm.Configure(Ringing).
        Permit(HungUp, OffHook).
        Permit(CallConnected, Connected)

    sm.Configure(Connected).
        OnEntry(func(ctx context.Context) error {
            fmt.Println("Call connected!")
            return nil
        }).
        Permit(HungUp, OffHook).
        Permit(PlacedOnHold, OnHold)

    sm.Configure(OnHold).
        SubstateOf(Connected).
        Permit(TakenOffHold, Connected).
        Permit(HungUp, OffHook)

    stateless.OnTransitioned[PhoneState, PhoneTrigger, stateless.NoArgs](sm,
        func(t stateless.Transition[PhoneState, PhoneTrigger, stateless.NoArgs]) {
            fmt.Printf("Transitioned: %v -> %v\n", t.Source, t.Destination)
        })

    // Generate graph
    info := sm.GetInfo()
    fmt.Println(graph.UmlDotGraph(info))

    // Use the state machine
    sm.Fire(CallDialed, nil)
    sm.Fire(CallConnected, nil)
    sm.Fire(PlacedOnHold, nil)
    sm.Fire(TakenOffHold, nil)
    sm.Fire(HungUp, nil)
}
```

## License

MIT License - see LICENSE file for details.

## Acknowledgments

This library is a Go port of the excellent [.NET Stateless library](https://github.com/dotnet-state-machine/stateless) by Nicholas Blumhardt and contributors.
