package stateless

import (
	"fmt"
	"reflect"
)

// TriggerWithParameters associates configured parameters with an underlying trigger value.
type TriggerWithParameters[TTrigger comparable] struct {
	underlyingTrigger TTrigger
	argumentTypes     []reflect.Type
}

// NewTriggerWithParameters creates a new configured trigger.
func NewTriggerWithParameters[TTrigger comparable](underlyingTrigger TTrigger, argumentTypes ...reflect.Type) *TriggerWithParameters[TTrigger] {
	return &TriggerWithParameters[TTrigger]{
		underlyingTrigger: underlyingTrigger,
		argumentTypes:     argumentTypes,
	}
}

// ArgumentTypes returns the argument types expected by this trigger.
func (t *TriggerWithParameters[TTrigger]) ArgumentTypes() []reflect.Type {
	return t.argumentTypes
}

// Trigger returns the underlying trigger value.
func (t *TriggerWithParameters[TTrigger]) Trigger() TTrigger {
	return t.underlyingTrigger
}

// ValidateParameters ensures that the supplied arguments are compatible with those configured for this trigger.
func (t *TriggerWithParameters[TTrigger]) ValidateParameters(args []any) error {
	if args == nil {
		return &ArgumentError{ParamName: "args", Message: "args cannot be nil"}
	}

	if len(args) > len(t.argumentTypes) {
		return &ParameterConversionError{
			Message: fmt.Sprintf("too many parameters have been supplied. Expected %d but got %d", len(t.argumentTypes), len(args)),
		}
	}

	for i, expectedType := range t.argumentTypes {
		if i >= len(args) {
			break
		}
		arg := args[i]
		if arg == nil {
			continue
		}
		argType := reflect.TypeOf(arg)
		if !argType.AssignableTo(expectedType) {
			return &ParameterConversionError{
				Message: fmt.Sprintf("argument at position %d is of type %v but expected type %v", i, argType, expectedType),
			}
		}
	}

	return nil
}

// TriggerWithParameters1 is a configured trigger with one required argument.
type TriggerWithParameters1[TTrigger comparable, TArg0 any] struct {
	*TriggerWithParameters[TTrigger]
}

// NewTriggerWithParameters1 creates a new configured trigger with one argument.
func NewTriggerWithParameters1[TTrigger comparable, TArg0 any](underlyingTrigger TTrigger) *TriggerWithParameters1[TTrigger, TArg0] {
	var zero TArg0
	return &TriggerWithParameters1[TTrigger, TArg0]{
		TriggerWithParameters: NewTriggerWithParameters(underlyingTrigger, reflect.TypeOf(zero)),
	}
}

// TriggerWithParameters2 is a configured trigger with two required arguments.
type TriggerWithParameters2[TTrigger comparable, TArg0, TArg1 any] struct {
	*TriggerWithParameters[TTrigger]
}

// NewTriggerWithParameters2 creates a new configured trigger with two arguments.
func NewTriggerWithParameters2[TTrigger comparable, TArg0, TArg1 any](underlyingTrigger TTrigger) *TriggerWithParameters2[TTrigger, TArg0, TArg1] {
	var zero0 TArg0
	var zero1 TArg1
	return &TriggerWithParameters2[TTrigger, TArg0, TArg1]{
		TriggerWithParameters: NewTriggerWithParameters(underlyingTrigger, reflect.TypeOf(zero0), reflect.TypeOf(zero1)),
	}
}

// TriggerWithParameters3 is a configured trigger with three required arguments.
type TriggerWithParameters3[TTrigger comparable, TArg0, TArg1, TArg2 any] struct {
	*TriggerWithParameters[TTrigger]
}

// NewTriggerWithParameters3 creates a new configured trigger with three arguments.
func NewTriggerWithParameters3[TTrigger comparable, TArg0, TArg1, TArg2 any](underlyingTrigger TTrigger) *TriggerWithParameters3[TTrigger, TArg0, TArg1, TArg2] {
	var zero0 TArg0
	var zero1 TArg1
	var zero2 TArg2
	return &TriggerWithParameters3[TTrigger, TArg0, TArg1, TArg2]{
		TriggerWithParameters: NewTriggerWithParameters(underlyingTrigger, reflect.TypeOf(zero0), reflect.TypeOf(zero1), reflect.TypeOf(zero2)),
	}
}

// TriggerDetails represents a trigger with details of any configured trigger parameters.
type TriggerDetails[TState, TTrigger comparable] struct {
	// Trigger is the trigger value.
	Trigger TTrigger

	// HasParameters indicates whether the trigger has been configured with parameters.
	HasParameters bool

	// Parameters contains the trigger parameter configuration, if any.
	Parameters *TriggerWithParameters[TTrigger]
}

// NewTriggerDetails creates a new TriggerDetails.
func NewTriggerDetails[TState, TTrigger comparable](
	trigger TTrigger,
	triggerConfiguration map[TTrigger]*TriggerWithParameters[TTrigger],
) TriggerDetails[TState, TTrigger] {
	params, hasParams := triggerConfiguration[trigger]
	return TriggerDetails[TState, TTrigger]{
		Trigger:       trigger,
		HasParameters: hasParams,
		Parameters:    params,
	}
}
