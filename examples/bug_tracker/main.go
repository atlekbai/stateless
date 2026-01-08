// Package main demonstrates a bug tracker state machine with guards and actions.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/atlekbai/stateless"
)

// State represents bug states.
type State int

const (
	Open State = iota
	Assigned
	InProgress
	Resolved
	Closed
	Reopened
)

func (s State) String() string {
	names := []string{"Open", "Assigned", "InProgress", "Resolved", "Closed", "Reopened"}
	if int(s) < len(names) {
		return names[s]
	}
	return "Unknown"
}

// Trigger represents bug triggers.
type Trigger int

const (
	Assign Trigger = iota
	StartWork
	Resolve
	Close
	Reopen
)

func (t Trigger) String() string {
	names := []string{"Assign", "StartWork", "Resolve", "Close", "Reopen"}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

// AssignArgs contains arguments for the Assign trigger.
type AssignArgs struct {
	Assignee string
}

type ResolveArgs struct {
	Resolution string
}

// Bug represents a bug in the tracker.
type Bug struct {
	ID         int
	Title      string
	Assignee   string
	Resolution string
	sm         *stateless.StateMachine[State, Trigger]
}

// NewBug creates a new bug with the state machine configured.
func NewBug(id int, title string) *Bug {
	bug := &Bug{
		ID:    id,
		Title: title,
	}

	// Create the state machine
	bug.sm = stateless.NewStateMachine[State, Trigger](Open)

	// Configure states
	bug.sm.Configure(Open).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			fmt.Printf("  Bug #%d is now open\n", bug.ID)
			return nil
		}).
		Permit(Assign, Assigned)

	// Configure Assigned state with typed entry action using type assertion
	bug.sm.Configure(Assigned).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			if args, ok := t.Args.(AssignArgs); ok {
				bug.Assignee = args.Assignee
				fmt.Printf("  Bug #%d assigned to: %s\n", bug.ID, args.Assignee)
			}
			return nil
		}).
		PermitReentry(Assign). // Can be reassigned
		Permit(StartWork, InProgress)

	bug.sm.Configure(InProgress).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			fmt.Printf("  Work started on bug #%d\n", bug.ID)
			return nil
		}).
		PermitIf(Resolve, Resolved, func() bool {
			return bug.Assignee != ""
		}, "Must have an assignee to resolve")

	// Configure Resolved state with typed entry action using type assertion
	bug.sm.Configure(Resolved).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			if args, ok := t.Args.(ResolveArgs); ok {
				bug.Resolution = args.Resolution
				fmt.Printf("  Bug #%d resolved: %s\n", bug.ID, args.Resolution)
			}
			return nil
		}).
		Permit(Close, Closed).
		Permit(Reopen, Reopened)

	bug.sm.Configure(Closed).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			fmt.Printf("  Bug #%d closed\n", bug.ID)
			return nil
		}).
		Permit(Reopen, Reopened)

	bug.sm.Configure(Reopened).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			bug.Resolution = ""
			fmt.Printf("  Bug #%d reopened\n", bug.ID)
			return nil
		}).
		Permit(Assign, Assigned)

	// Set up event handlers
	bug.sm.OnTransitioned(func(t stateless.Transition[State, Trigger]) {
		fmt.Printf("  [Transition] %s -> %s (trigger: %s)\n", t.Source, t.Destination, t.Trigger)
	})

	return bug
}

func (b *Bug) State() State {
	return b.sm.State()
}

func (b *Bug) Fire(trigger Trigger, args any) error {
	return b.sm.Fire(trigger, args)
}

func (b *Bug) CanFire(trigger Trigger) bool {
	return b.sm.CanFire(trigger, nil)
}

func main() {
	fmt.Println("Bug Tracker State Machine Example")
	fmt.Println("==================================")
	fmt.Println()

	// Create a new bug
	bug := NewBug(123, "Login button doesn't work")
	fmt.Printf("Created bug #%d: %s\n", bug.ID, bug.Title)
	fmt.Printf("State: %s\n\n", bug.State())

	// Assign the bug
	fmt.Println("Assigning bug...")
	if err := bug.Fire(Assign, AssignArgs{Assignee: "Alice"}); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Start work
	fmt.Println("Starting work...")
	if err := bug.Fire(StartWork, nil); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Try to resolve
	fmt.Println("Resolving bug...")
	if err := bug.Fire(Resolve, ResolveArgs{Resolution: "Fixed the CSS issue"}); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Close the bug
	fmt.Println("Closing bug...")
	if err := bug.Fire(Close, nil); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Reopen the bug
	fmt.Println("Reopening bug...")
	if err := bug.Fire(Reopen, nil); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	fmt.Println("Final state:", bug.State())
}
