// Package main demonstrates a phone call state machine example.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/atlekbai/stateless"
)

// State represents phone call states.
type State int

const (
	OffHook State = iota
	Ringing
	Connected
	OnHold
)

func (s State) String() string {
	switch s {
	case OffHook:
		return "OffHook"
	case Ringing:
		return "Ringing"
	case Connected:
		return "Connected"
	case OnHold:
		return "OnHold"
	default:
		return "Unknown"
	}
}

// Trigger represents phone call triggers.
type Trigger int

const (
	CallDialed Trigger = iota
	CallConnected
	PlacedOnHold
	TakenOffHold
	LeftMessage
	HungUp
)

func (t Trigger) String() string {
	switch t {
	case CallDialed:
		return "CallDialed"
	case CallConnected:
		return "CallConnected"
	case PlacedOnHold:
		return "PlacedOnHold"
	case TakenOffHold:
		return "TakenOffHold"
	case LeftMessage:
		return "LeftMessage"
	case HungUp:
		return "HungUp"
	default:
		return "Unknown"
	}
}

func main() {
	fmt.Println("Phone Call State Machine Example")
	fmt.Println("================================")

	// Create a new state machine starting in OffHook state
	sm := stateless.NewStateMachine[State, Trigger](OffHook)

	// Configure the state machine
	sm.Configure(OffHook).
		Permit(CallDialed, Ringing)

	sm.Configure(Ringing).
		Permit(HungUp, OffHook).
		Permit(CallConnected, Connected)

	sm.Configure(Connected).
		OnEntry(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			fmt.Println("  -> Call connected!")
			return nil
		}).
		OnExit(func(ctx context.Context, t stateless.Transition[State, Trigger]) error {
			fmt.Println("  -> Call ended.")
			return nil
		}).
		Permit(LeftMessage, OffHook).
		Permit(HungUp, OffHook).
		Permit(PlacedOnHold, OnHold)

	sm.Configure(OnHold).
		SubstateOf(Connected).
		Permit(TakenOffHold, Connected).
		Permit(HungUp, OffHook)

	// Subscribe to transitions
	sm.OnTransitioned(func(t stateless.Transition[State, Trigger]) {
		fmt.Printf("  Transitioned from %s to %s via %s\n", t.Source, t.Destination, t.Trigger)
	})

	// Fire some triggers
	printState(sm)

	fire(sm, CallDialed)
	printState(sm)

	fire(sm, CallConnected)
	printState(sm)

	fire(sm, PlacedOnHold)
	printState(sm)

	fire(sm, TakenOffHold)
	printState(sm)

	fire(sm, HungUp)
	printState(sm)
}

func fire(sm *stateless.StateMachine[State, Trigger], trigger Trigger) {
	fmt.Printf("Firing trigger: %s\n", trigger)
	if err := sm.Fire(trigger, nil); err != nil {
		log.Printf("Error firing trigger %s: %v", trigger, err)
	}
}

func printState(sm *stateless.StateMachine[State, Trigger]) {
	fmt.Printf("Current state: %s\n", sm.State())
	fmt.Printf("Permitted triggers: %v\n\n", sm.GetPermittedTriggers(nil))
}
