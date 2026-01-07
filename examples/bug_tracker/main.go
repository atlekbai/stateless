// Package main demonstrates a bug tracker state machine with guards and actions.
package main

import (
	"fmt"
	"log"

	"github.com/atlekbai/stateless"
)

// Bug states
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

// Bug triggers
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

// Argument structs for typed transitions
type AssignArgs struct {
	Assignee string
}

type ResolveArgs struct {
	Resolution string
}

// Bug represents a bug in the tracker
type Bug struct {
	ID         int
	Title      string
	Assignee   string
	Resolution string
	sm         *stateless.StateMachine[State, Trigger]
}

// NewBug creates a new bug with the state machine configured
func NewBug(id int, title string) *Bug {
	bug := &Bug{
		ID:    id,
		Title: title,
	}

	// Create the state machine
	bug.sm = stateless.NewStateMachine[State, Trigger](Open)

	// Configure states
	bug.sm.Configure(Open).
		OnEntryAction(func() { fmt.Printf("  Bug #%d is now open\n", bug.ID) }).
		Permit(Assign, Assigned)

	// Configure Assigned state with typed entry action
	assignedConfig := bug.sm.Configure(Assigned).
		PermitReentry(Assign). // Can be reassigned
		Permit(StartWork, InProgress)

	stateless.OnEntry[State, Trigger, AssignArgs](assignedConfig, func(t stateless.Transition[State, Trigger, AssignArgs]) {
		bug.Assignee = t.Args.Assignee
		fmt.Printf("  Bug #%d assigned to: %s\n", bug.ID, t.Args.Assignee)
	})

	bug.sm.Configure(InProgress).
		OnEntryAction(func() { fmt.Printf("  Work started on bug #%d\n", bug.ID) }).
		PermitIf(Resolve, Resolved, func() bool {
			return bug.Assignee != ""
		}, "Must have an assignee to resolve")

	// Configure Resolved state with typed entry action
	resolvedConfig := bug.sm.Configure(Resolved).
		Permit(Close, Closed).
		Permit(Reopen, Reopened)

	stateless.OnEntry[State, Trigger, ResolveArgs](resolvedConfig, func(t stateless.Transition[State, Trigger, ResolveArgs]) {
		bug.Resolution = t.Args.Resolution
		fmt.Printf("  Bug #%d resolved: %s\n", bug.ID, t.Args.Resolution)
	})

	bug.sm.Configure(Closed).
		OnEntryAction(func() { fmt.Printf("  Bug #%d closed\n", bug.ID) }).
		Permit(Reopen, Reopened)

	bug.sm.Configure(Reopened).
		OnEntryAction(func() {
			bug.Resolution = ""
			fmt.Printf("  Bug #%d reopened\n", bug.ID)
		}).
		Permit(Assign, Assigned)

	// Set up event handlers
	stateless.OnTransitioned[State, Trigger, stateless.NoArgs](bug.sm, func(t stateless.Transition[State, Trigger, stateless.NoArgs]) {
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
