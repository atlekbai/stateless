package stateless_test

import (
	"testing"

	"github.com/atlekbai/stateless"
)

func TestTransition_IsReentry(t *testing.T) {
	trans := stateless.Transition[State, Trigger]{Source: StateA, Destination: StateA, Trigger: TriggerX}
	if !trans.IsReentry() {
		t.Error("expected IsReentry to be true for same source and destination")
	}

	trans2 := stateless.Transition[State, Trigger]{Source: StateA, Destination: StateB, Trigger: TriggerX}
	if trans2.IsReentry() {
		t.Error("expected IsReentry to be false for different source and destination")
	}
}
