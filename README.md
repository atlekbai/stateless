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
        OnExit(func() { fmt.Println("Light turning on...") })

    sm.Configure(On).
        Permit(Toggle, Off).
        OnEntry(func() { fmt.Println("Light is on!") })

    // Fire triggers
    sm.Fire(Toggle) // Off -> On
    sm.Fire(Toggle) // On -> Off
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
    InternalTransition(TriggerX, func(t stateless.Transition[State, Trigger], args ...any) {
        // Action executed without state change
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
    OnEntry(func() {
        fmt.Println("Entering StateA")
    }).
    OnExit(func() {
        fmt.Println("Exiting StateA")
    }).
    OnEntryFrom(TriggerX, func() {
        fmt.Println("Entered from TriggerX")
    }).
    OnEntryWithTransition(func(t stateless.Transition[State, Trigger]) {
        fmt.Printf("Entered from %v\n", t.Source)
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
stateless.OnEntryWithTransition[State, Trigger, CallArgs](sm.Configure(StateB),
    func(t stateless.Transition[State, Trigger, CallArgs]) {
        fmt.Printf("Call from: %s\n", t.Args.CallerID)
    })

// Fire with struct argument
sm.Fire(TriggerX, CallArgs{CallerID: "555-1234"})
```

## Activation and Deactivation

```go
sm.Configure(StateA).
    OnActivate(func() {
        fmt.Println("State machine activated in StateA")
    }).
    OnDeactivate(func() {
        fmt.Println("State machine deactivated in StateA")
    })

sm.Activate()    // Calls OnActivate
sm.Deactivate()  // Calls OnDeactivate
```

## Event Handlers

```go
// Called when a transition occurs
sm.OnTransitioned(func(t stateless.Transition[State, Trigger]) {
    fmt.Printf("%v -> %v\n", t.Source, t.Destination)
})

// Called after all transition actions complete
sm.OnTransitionCompleted(func(t stateless.Transition[State, Trigger]) {
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
triggers := sm.GetPermittedTriggers()

// Get detailed trigger info
details := sm.GetDetailedPermittedTriggers()

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

```go
import "context"

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := sm.FireCtx(ctx, TriggerX)
```

## Complete Example

```go
package main

import (
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
        OnEntry(func() { fmt.Println("Call connected!") }).
        Permit(HungUp, OffHook).
        Permit(PlacedOnHold, OnHold)

    sm.Configure(OnHold).
        SubstateOf(Connected).
        Permit(TakenOffHold, Connected).
        Permit(HungUp, OffHook)

    sm.OnTransitioned(func(t stateless.Transition[PhoneState, PhoneTrigger]) {
        fmt.Printf("Transitioned: %v -> %v\n", t.Source, t.Destination)
    })

    // Generate graph
    info := sm.GetInfo()
    fmt.Println(graph.UmlDotGraph(info))

    // Use the state machine
    sm.Fire(CallDialed)
    sm.Fire(CallConnected)
    sm.Fire(PlacedOnHold)
    sm.Fire(TakenOffHold)
    sm.Fire(HungUp)
}
```

## License

MIT License - see LICENSE file for details.

## Acknowledgments

This library is a Go port of the excellent [.NET Stateless library](https://github.com/dotnet-state-machine/stateless) by Nicholas Blumhardt and contributors.
