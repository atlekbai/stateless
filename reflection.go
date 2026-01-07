package stateless

import (
	"reflect"
	"runtime"
	"strings"
)

// Timing indicates whether a method is synchronous or asynchronous.
type Timing int

const (
	// TimingSynchronous indicates the method is synchronous.
	TimingSynchronous Timing = iota
	// TimingAsynchronous indicates the method is asynchronous.
	TimingAsynchronous
)

// InvocationInfo describes a method - either an action or a guard condition.
type InvocationInfo struct {
	// MethodName is the name of the invoked method.
	MethodName string
	// description is the user-specified description (can be empty).
	description string
	// timing indicates if the method is synchronous or asynchronous.
	timing Timing
}

// DefaultFunctionDescription is the text returned for compiler-generated functions
// where the caller has not specified a description.
var DefaultFunctionDescription = "Function"

// NullString is the string representation of a null value.
const NullString = "<null>"

// NewInvocationInfo creates a new InvocationInfo.
func NewInvocationInfo(methodName, description string, timing Timing) InvocationInfo {
	return InvocationInfo{
		MethodName:  methodName,
		description: description,
		timing:      timing,
	}
}

// CreateInvocationInfo creates InvocationInfo from a function and description.
func CreateInvocationInfo(fn any, description string, timing Timing) InvocationInfo {
	methodName := getFunctionName(fn)
	return NewInvocationInfo(methodName, description, timing)
}

// Description returns the description of the invoked method.
// Returns:
// 1. The user-specified description, if any
// 2. Otherwise, if the method name is compiler-generated, returns DefaultFunctionDescription
// 3. Otherwise, the method name
func (i InvocationInfo) Description() string {
	if i.description != "" {
		return i.description
	}
	if i.MethodName == "" {
		return NullString
	}
	// Check for anonymous/compiler-generated function names
	if strings.Contains(i.MethodName, "func") || strings.Contains(i.MethodName, ".") {
		return DefaultFunctionDescription
	}
	return i.MethodName
}

// IsAsync returns true if the method is invoked asynchronously.
func (i InvocationInfo) IsAsync() bool {
	return i.timing == TimingAsynchronous
}

// getFunctionName returns the name of a function.
func getFunctionName(fn any) string {
	if fn == nil {
		return ""
	}
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	// Extract just the function name from the full path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

// ActionInfo describes an action with optional trigger information.
type ActionInfo struct {
	InvocationInfo
	// FromTrigger is the trigger that causes this action to execute (optional).
	FromTrigger any
}

// NewActionInfo creates a new ActionInfo.
func NewActionInfo(method InvocationInfo, fromTrigger any) ActionInfo {
	return ActionInfo{
		InvocationInfo: method,
		FromTrigger:    fromTrigger,
	}
}

// TriggerInfo describes a trigger.
type TriggerInfo struct {
	// UnderlyingTrigger is the underlying trigger value.
	UnderlyingTrigger any
}

// NewTriggerInfo creates a new TriggerInfo.
func NewTriggerInfo(trigger any) TriggerInfo {
	return TriggerInfo{UnderlyingTrigger: trigger}
}

// String returns the string representation of the trigger.
func (t TriggerInfo) String() string {
	if t.UnderlyingTrigger == nil {
		return NullString
	}
	if s, ok := t.UnderlyingTrigger.(interface{ String() string }); ok {
		return s.String()
	}
	return ""
}

// StateMachineInfo exposes the states, transitions, and actions of a state machine.
type StateMachineInfo struct {
	// InitialState is the initial state of the state machine.
	InitialState *StateInfo

	// States contains all states in the machine.
	States []*StateInfo

	// StateType is a string representation of the state type.
	StateType string

	// TriggerType is a string representation of the trigger type.
	TriggerType string
}

// StateInfo describes an internal state representation through the reflection API.
type StateInfo struct {
	// UnderlyingState is the instance or value this state represents.
	UnderlyingState any

	// Superstate is the superstate defined, if any.
	Superstate *StateInfo

	// Substates are substates defined for this state.
	Substates []*StateInfo

	// EntryActions are actions executed on state-entry.
	EntryActions []ActionInfo

	// ActivateActions are actions executed on activation.
	ActivateActions []InvocationInfo

	// DeactivateActions are actions executed on deactivation.
	DeactivateActions []InvocationInfo

	// ExitActions are actions executed on state-exit.
	ExitActions []InvocationInfo

	// FixedTransitions are fixed transitions defined for this state.
	FixedTransitions []FixedTransitionInfo

	// DynamicTransitions are dynamic transitions defined for this state.
	DynamicTransitions []DynamicTransitionInfo

	// IgnoredTriggers are triggers ignored for this state.
	IgnoredTriggers []IgnoredTransitionInfo
}

// String returns the string representation of the state.
func (s *StateInfo) String() string {
	if s == nil || s.UnderlyingState == nil {
		return NullString
	}
	if str, ok := s.UnderlyingState.(interface{ String() string }); ok {
		return str.String()
	}
	return ""
}

// Transitions returns all transitions (both fixed and dynamic) defined for this state.
func (s *StateInfo) Transitions() []TransitionInfo {
	result := make([]TransitionInfo, 0, len(s.FixedTransitions)+len(s.DynamicTransitions))
	for i := range s.FixedTransitions {
		result = append(result, &s.FixedTransitions[i])
	}
	for i := range s.DynamicTransitions {
		result = append(result, &s.DynamicTransitions[i])
	}
	return result
}

// TransitionInfo is the base interface for transition information.
type TransitionInfo interface {
	// GetTrigger returns the trigger that causes this transition.
	GetTrigger() TriggerInfo
	// GetGuardConditions returns the guard conditions for this transition.
	GetGuardConditions() []InvocationInfo
	// GetIsInternalTransition returns true if this is an internal transition.
	GetIsInternalTransition() bool
}

// transitionInfoBase contains common fields for transition information.
type transitionInfoBase struct {
	// Trigger is the trigger whose firing resulted in this transition.
	Trigger TriggerInfo

	// GuardConditions contains method descriptions of the guard conditions.
	GuardConditions []InvocationInfo

	// IsInternalTransition indicates if this is an internal transition.
	IsInternalTransition bool
}

func (t *transitionInfoBase) GetTrigger() TriggerInfo {
	return t.Trigger
}

func (t *transitionInfoBase) GetGuardConditions() []InvocationInfo {
	return t.GuardConditions
}

func (t *transitionInfoBase) GetIsInternalTransition() bool {
	return t.IsInternalTransition
}

// FixedTransitionInfo describes a transition that can be initiated from a trigger.
type FixedTransitionInfo struct {
	transitionInfoBase

	// DestinationState is the state that will be transitioned into on activation.
	DestinationState *StateInfo
}

// DynamicStateInfo contains information about a possible destination state for a dynamic transition.
type DynamicStateInfo struct {
	// DestinationState is the name of the destination state.
	DestinationState string

	// Criterion is the reason this destination state was chosen.
	Criterion string
}

// DynamicTransitionInfo describes a transition that can be initiated from a trigger,
// but whose result is non-deterministic.
type DynamicTransitionInfo struct {
	transitionInfoBase

	// DestinationStateSelectorDescription is the method information for the destination state selector.
	DestinationStateSelectorDescription InvocationInfo

	// PossibleDestinationStates are the possible destination states.
	PossibleDestinationStates []DynamicStateInfo
}

// IgnoredTransitionInfo describes a trigger that is ignored in a state.
type IgnoredTransitionInfo struct {
	transitionInfoBase
}
