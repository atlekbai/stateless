package stateless

import (
	"fmt"
	"strings"
)

// InvalidOperationError indicates an operation that is not valid given the current state.
type InvalidOperationError struct {
	Message string
}

func (e *InvalidOperationError) Error() string {
	return e.Message
}

// ArgumentError indicates an invalid argument was passed.
type ArgumentError struct {
	ParamName string
	Message   string
}

func (e *ArgumentError) Error() string {
	if e.ParamName != "" {
		return fmt.Sprintf("%s (parameter: %s)", e.Message, e.ParamName)
	}
	return e.Message
}

// InvalidTransitionError is thrown when a trigger is fired from a state that
// does not have a valid transition for that trigger.
type InvalidTransitionError struct {
	Trigger           any
	State             any
	UnmetGuards       []error
	PermittedTriggers []any
}

func (e *InvalidTransitionError) Error() string {
	if len(e.UnmetGuards) > 0 {
		guardMessages := make([]string, len(e.UnmetGuards))
		for i, err := range e.UnmetGuards {
			guardMessages[i] = err.Error()
		}
		return fmt.Sprintf(
			"trigger '%v' is valid for transition from state '%v' "+
				"but guard conditions are not met. Guard conditions: %s",
			e.Trigger, e.State, strings.Join(guardMessages, ", "))
	}

	var permitted string
	if len(e.PermittedTriggers) > 0 {
		triggers := make([]string, len(e.PermittedTriggers))
		for i, t := range e.PermittedTriggers {
			triggers[i] = fmt.Sprintf("%v", t)
		}
		permitted = fmt.Sprintf(" Permitted triggers: %s.", strings.Join(triggers, ", "))
	} else {
		permitted = " No valid leaving transitions are permitted from state."
	}

	return fmt.Sprintf(
		"no valid leaving transitions are permitted from state '%v' for trigger '%v'.%s",
		e.State, e.Trigger, permitted)
}

// ParameterConversionError indicates an error during parameter conversion.
type ParameterConversionError struct {
	Message string
}

func (e *ParameterConversionError) Error() string {
	return e.Message
}
