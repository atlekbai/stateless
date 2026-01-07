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

// Bug represents a bug in the tracker
type Bug struct {
	ID          int
	Title       string
	Assignee    string
	Resolution  string
	sm          *stateless.StateMachine[State, Trigger]
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
		OnEntry(func() { fmt.Printf("  Bug #%d is now open\n", bug.ID) }).
		Permit(Assign, Assigned)

	bug.sm.Configure(Assigned).
		OnEntryWithArgs(func(t stateless.Transition[State, Trigger], args ...any) {
			if len(args) > 0 {
				if assignee, ok := args[0].(string); ok {
					bug.Assignee = assignee
					fmt.Printf("  Bug #%d assigned to: %s\n", bug.ID, assignee)
				}
			}
		}).
		PermitReentry(Assign). // Can be reassigned
		Permit(StartWork, InProgress)

	bug.sm.Configure(InProgress).
		OnEntry(func() { fmt.Printf("  Work started on bug #%d\n", bug.ID) }).
		PermitIf(Resolve, Resolved, func() bool {
			return bug.Assignee != ""
		}, "Must have an assignee to resolve")

	bug.sm.Configure(Resolved).
		OnEntryWithArgs(func(t stateless.Transition[State, Trigger], args ...any) {
			if len(args) > 0 {
				if resolution, ok := args[0].(string); ok {
					bug.Resolution = resolution
					fmt.Printf("  Bug #%d resolved: %s\n", bug.ID, resolution)
				}
			}
		}).
		Permit(Close, Closed).
		Permit(Reopen, Reopened)

	bug.sm.Configure(Closed).
		OnEntry(func() { fmt.Printf("  Bug #%d closed\n", bug.ID) }).
		Permit(Reopen, Reopened)

	bug.sm.Configure(Reopened).
		OnEntry(func() {
			bug.Resolution = ""
			fmt.Printf("  Bug #%d reopened\n", bug.ID)
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

func (b *Bug) Fire(trigger Trigger, args ...any) error {
	return b.sm.Fire(trigger, args...)
}

func (b *Bug) CanFire(trigger Trigger) bool {
	return b.sm.CanFire(trigger)
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
	if err := bug.Fire(Assign, "Alice"); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Start work
	fmt.Println("Starting work...")
	if err := bug.Fire(StartWork); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Try to resolve
	fmt.Println("Resolving bug...")
	if err := bug.Fire(Resolve, "Fixed the CSS issue"); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Close the bug
	fmt.Println("Closing bug...")
	if err := bug.Fire(Close); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	// Reopen the bug
	fmt.Println("Reopening bug...")
	if err := bug.Fire(Reopen); err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("State: %s\n\n", bug.State())

	fmt.Println("Final state:", bug.State())
}
