// Package stateless provides a feature-complete, generic state machine library for Go.
//
// This library is a Go port of the .NET Stateless library, providing a fluent API
// for configuring state machines with support for:
//
//   - Generic types for states and triggers
//   - Guard conditions for conditional transitions
//   - Entry and exit actions
//   - Hierarchical states (substates and superstates)
//   - Parameterized triggers
//   - Dynamic transitions
//   - Reentry and internal transitions
//   - Firing modes (immediate or queued)
//   - Introspection and graph generation
//
// # Basic Usage
//
// Create a state machine with initial state:
//
//	sm := stateless.NewStateMachine[State, Trigger](InitialState)
//
// Configure states with transitions:
//
//	sm.Configure(StateA).
//	    Permit(TriggerX, StateB).
//	    OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
//	        fmt.Println("Entering StateA")
//	        return nil
//	    })
//
// Fire triggers to cause transitions:
//
//	err := sm.Fire(TriggerX, nil)
//
// # Guards
//
// Add conditions to transitions. Guards return nil to allow the transition,
// or an error describing why it's blocked:
//
//	sm.Configure(StateA).
//	    PermitIf(TriggerX, StateB, func(args any) error {
//	        if !someCondition {
//	            return errors.New("condition not met")
//	        }
//	        return nil
//	    })
//
// # Hierarchical States
//
// Create state hierarchies:
//
//	sm.Configure(StateB).SubstateOf(StateA)
//
// # Graph Generation
//
// Export to DOT or Mermaid format:
//
//	import "github.com/atlekbai/stateless/graph"
//	dot := graph.UmlDotGraph(sm.GetInfo())
package stateless
