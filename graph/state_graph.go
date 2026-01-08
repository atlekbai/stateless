package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/atlekbai/stateless"
)

// StateGraph generates a symbolic representation of the graph structure.
type StateGraph struct {
	// InitialState is the initial state of the machine.
	InitialState *stateless.StateInfo

	// States contains all states in the graph, indexed by state name.
	States map[string]*State

	// Transitions contains all transitions in the graph.
	Transitions []*Transition

	// Decisions contains all decision nodes in the graph (for dynamic transitions).
	Decisions []*Decision
}

// NewStateGraph creates a new state graph from state machine info.
func NewStateGraph(machineInfo *stateless.StateMachineInfo) *StateGraph {
	sg := &StateGraph{
		InitialState: machineInfo.InitialState,
		States:       make(map[string]*State),
	}

	// Add superstates first
	sg.addSuperstates(machineInfo)

	// Add single states (states that aren't part of a hierarchy)
	sg.addSingleStates(machineInfo)

	// Add transitions
	sg.addTransitions(machineInfo)

	// Process OnEntryFrom actions
	sg.processOnEntryFrom(machineInfo)

	return sg
}

// addSuperstates adds superstates to the graph.
func (sg *StateGraph) addSuperstates(machineInfo *stateless.StateMachineInfo) {
	for _, stateInfo := range machineInfo.States {
		if len(stateInfo.Substates) > 0 && stateInfo.Superstate == nil {
			state := sg.createSuperState(stateInfo)
			sg.States[fmt.Sprintf("%v", stateInfo.UnderlyingState)] = state.State
			sg.addSubstates(state, stateInfo.Substates)
		}
	}
}

// createSuperState creates a SuperState from StateInfo.
func (sg *StateGraph) createSuperState(stateInfo *stateless.StateInfo) *SuperState {
	state := &State{
		StateName:    fmt.Sprintf("%v", stateInfo.UnderlyingState),
		NodeName:     fmt.Sprintf("%v", stateInfo.UnderlyingState),
		EntryActions: sg.extractEntryActionDescriptions(stateInfo),
		ExitActions:  sg.extractExitActionDescriptions(stateInfo),
		StateInfo:    stateInfo,
	}
	return &SuperState{
		State:     state,
		SubStates: make([]*State, 0),
	}
}

// addSubstates recursively adds substates to a superstate.
func (sg *StateGraph) addSubstates(superState *SuperState, substates []*stateless.StateInfo) {
	for _, subStateInfo := range substates {
		stateName := fmt.Sprintf("%v", subStateInfo.UnderlyingState)
		if _, exists := sg.States[stateName]; exists {
			continue
		}

		if len(subStateInfo.Substates) > 0 {
			// This is also a superstate
			sub := sg.createSuperState(subStateInfo)
			sg.States[stateName] = sub.State
			superState.SubStates = append(superState.SubStates, sub.State)
			sub.State.SuperState = superState
			sg.addSubstates(sub, subStateInfo.Substates)
		} else {
			// Regular state
			sub := &State{
				StateName:    stateName,
				NodeName:     stateName,
				EntryActions: sg.extractEntryActionDescriptions(subStateInfo),
				ExitActions:  sg.extractExitActionDescriptions(subStateInfo),
				StateInfo:    subStateInfo,
			}
			sg.States[stateName] = sub
			superState.SubStates = append(superState.SubStates, sub)
			sub.SuperState = superState
		}
	}
}

// addSingleStates adds states that aren't part of a hierarchy.
func (sg *StateGraph) addSingleStates(machineInfo *stateless.StateMachineInfo) {
	for _, stateInfo := range machineInfo.States {
		stateName := fmt.Sprintf("%v", stateInfo.UnderlyingState)
		if _, exists := sg.States[stateName]; !exists {
			sg.States[stateName] = &State{
				StateName:    stateName,
				NodeName:     stateName,
				EntryActions: sg.extractEntryActionDescriptions(stateInfo),
				ExitActions:  sg.extractExitActionDescriptions(stateInfo),
				StateInfo:    stateInfo,
			}
		}
	}
}

// addTransitions adds all transitions to the graph.
func (sg *StateGraph) addTransitions(machineInfo *stateless.StateMachineInfo) {
	for _, stateInfo := range machineInfo.States {
		fromStateName := fmt.Sprintf("%v", stateInfo.UnderlyingState)
		fromState := sg.States[fromStateName]

		// Add fixed transitions
		for _, fix := range stateInfo.FixedTransitions {
			toStateName := fmt.Sprintf("%v", fix.DestinationState.UnderlyingState)
			toState := sg.States[toStateName]

			if fromState == toState {
				// Stay transition (self-loop)
				stay := &StayTransition{
					Transition: &Transition{
						Trigger:                 fix.GetTrigger(),
						SourceState:             fromState,
						DestinationState:        toState,
						Guards:                  fix.GetGuardConditions(),
						ExecuteEntryExitActions: !fix.GetIsInternalTransition(),
					},
				}
				sg.Transitions = append(sg.Transitions, stay.Transition)
				fromState.Leaving = append(fromState.Leaving, stay.Transition)
				fromState.Arriving = append(fromState.Arriving, stay.Transition)

				// Add entry actions if this is a reentry
				if stay.ExecuteEntryExitActions {
					for _, action := range stateInfo.EntryActions {
						if action.FromTrigger == nil {
							stay.DestinationEntryActions = append(stay.DestinationEntryActions, action)
						}
					}
				}
			} else {
				// Regular transition
				trans := &FixedTransition{
					Transition: &Transition{
						Trigger:                 fix.GetTrigger(),
						SourceState:             fromState,
						DestinationState:        toState,
						Guards:                  fix.GetGuardConditions(),
						ExecuteEntryExitActions: true,
					},
				}
				sg.Transitions = append(sg.Transitions, trans.Transition)
				fromState.Leaving = append(fromState.Leaving, trans.Transition)
				toState.Arriving = append(toState.Arriving, trans.Transition)
			}
		}

		// Add dynamic transitions
		for _, dyn := range stateInfo.DynamicTransitions {
			// Create a decision node
			decide := &Decision{
				NodeName: fmt.Sprintf("Decision%d", len(sg.Decisions)+1),
				Method:   dyn.DestinationStateSelectorDescription,
			}
			sg.Decisions = append(sg.Decisions, decide)

			// Add transition from state to decision node
			trans := &FixedTransition{
				Transition: &Transition{
					Trigger:                 dyn.GetTrigger(),
					SourceState:             fromState,
					Guards:                  dyn.GetGuardConditions(),
					ExecuteEntryExitActions: true,
				},
			}
			sg.Transitions = append(sg.Transitions, trans.Transition)
			fromState.Leaving = append(fromState.Leaving, trans.Transition)
			decide.Arriving = append(decide.Arriving, trans.Transition)

			// Add transitions from decision node to possible destinations
			for _, possibleDest := range dyn.PossibleDestinationStates {
				if toState, exists := sg.States[possibleDest.DestinationState]; exists {
					trans := &Transition{
						Trigger:                 dyn.GetTrigger(),
						SourceState:             fromState,
						DestinationState:        toState,
						ExecuteEntryExitActions: true,
					}
					sg.Transitions = append(sg.Transitions, trans)
					decide.Leaving = append(decide.Leaving, trans)
					toState.Arriving = append(toState.Arriving, trans)
				}
			}
		}

		// Add ignored triggers
		for _, ignored := range stateInfo.IgnoredTriggers {
			stay := &StayTransition{
				Transition: &Transition{
					Trigger:                 ignored.GetTrigger(),
					SourceState:             fromState,
					DestinationState:        fromState,
					Guards:                  ignored.GetGuardConditions(),
					ExecuteEntryExitActions: false,
				},
			}
			sg.Transitions = append(sg.Transitions, stay.Transition)
			fromState.Leaving = append(fromState.Leaving, stay.Transition)
			fromState.Arriving = append(fromState.Arriving, stay.Transition)
		}
	}
}

// processOnEntryFrom processes entry actions that are bound to specific triggers.
func (sg *StateGraph) processOnEntryFrom(machineInfo *stateless.StateMachineInfo) {
	for _, stateInfo := range machineInfo.States {
		stateName := fmt.Sprintf("%v", stateInfo.UnderlyingState)
		state := sg.States[stateName]

		for _, entryAction := range stateInfo.EntryActions {
			if entryAction.FromTrigger != nil {
				// Find incoming transitions with this trigger
				for _, transit := range state.Arriving {
					if transit.ExecuteEntryExitActions {
						triggerStr := fmt.Sprintf("%v", transit.Trigger.UnderlyingTrigger)
						fromTriggerStr := fmt.Sprintf("%v", entryAction.FromTrigger)
						if triggerStr == fromTriggerStr {
							transit.DestinationEntryActions = append(transit.DestinationEntryActions, entryAction)
						}
					}
				}
			}
		}
	}
}

// extractEntryActionDescriptions extracts entry action descriptions from state info.
func (sg *StateGraph) extractEntryActionDescriptions(stateInfo *stateless.StateInfo) []string {
	var descriptions []string
	for _, action := range stateInfo.EntryActions {
		if action.FromTrigger == nil {
			descriptions = append(descriptions, action.Description())
		}
	}
	return descriptions
}

// extractExitActionDescriptions extracts exit action descriptions from state info.
func (sg *StateGraph) extractExitActionDescriptions(stateInfo *stateless.StateInfo) []string {
	var descriptions []string
	for _, action := range stateInfo.ExitActions {
		descriptions = append(descriptions, action.Description())
	}
	return descriptions
}

// ToGraph converts the state graph to a string representation using the specified style.
func (sg *StateGraph) ToGraph(style Style) string {
	var sb strings.Builder

	sb.WriteString(style.GetPrefix())

	// Get sorted state names for deterministic output
	sortedStateNames := sg.getSortedStateNames()

	// Format clusters (superstates) in sorted order
	for _, stateName := range sortedStateNames {
		state := sg.States[stateName]
		if superState, ok := sg.isSuperState(state); ok {
			sb.WriteString(style.FormatOneCluster(superState))
		}
	}

	// Format regular states (not superstates and not substates) in sorted order
	for _, stateName := range sortedStateNames {
		state := sg.States[stateName]
		if _, ok := sg.isSuperState(state); ok {
			continue
		}
		if sg.isDecision(state) || state.SuperState != nil {
			continue
		}
		sb.WriteString(style.FormatOneState(state))
	}

	// Format decision nodes
	for _, dec := range sg.Decisions {
		sb.WriteString(style.FormatOneDecisionNode(dec.NodeName, dec.Method.Description()))
	}

	// Sort transitions for deterministic output
	sortedTransitions := sg.getSortedTransitions()

	// Format transitions
	lines := style.FormatAllTransitions(sortedTransitions, sg.Decisions)
	for _, line := range lines {
		sb.WriteString("\n")
		sb.WriteString(line)
	}

	// Add initial transition
	sb.WriteString(style.GetInitialTransition(sg.InitialState))

	return sb.String()
}

// getSortedStateNames returns state names in sorted order for deterministic output.
func (sg *StateGraph) getSortedStateNames() []string {
	names := make([]string, 0, len(sg.States))
	for name := range sg.States {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getSortedTransitions returns transitions sorted by source state, then destination state, then trigger.
func (sg *StateGraph) getSortedTransitions() []*Transition {
	sorted := make([]*Transition, len(sg.Transitions))
	copy(sorted, sg.Transitions)
	sort.Slice(sorted, func(i, j int) bool {
		ti, tj := sorted[i], sorted[j]
		// Sort by source state name
		srcI := ""
		srcJ := ""
		if ti.SourceState != nil {
			srcI = ti.SourceState.StateName
		}
		if tj.SourceState != nil {
			srcJ = tj.SourceState.StateName
		}
		if srcI != srcJ {
			return srcI < srcJ
		}
		// Then by destination state name
		dstI := ""
		dstJ := ""
		if ti.DestinationState != nil {
			dstI = ti.DestinationState.StateName
		}
		if tj.DestinationState != nil {
			dstJ = tj.DestinationState.StateName
		}
		if dstI != dstJ {
			return dstI < dstJ
		}
		// Then by trigger
		trigI := fmt.Sprintf("%v", ti.Trigger.UnderlyingTrigger)
		trigJ := fmt.Sprintf("%v", tj.Trigger.UnderlyingTrigger)
		return trigI < trigJ
	})
	return sorted
}

// isSuperState checks if a state is a superstate.
func (sg *StateGraph) isSuperState(state *State) (*SuperState, bool) {
	if state.StateInfo != nil && len(state.StateInfo.Substates) > 0 {
		return &SuperState{
			State:     state,
			SubStates: sg.getSubStates(state),
		}, true
	}
	return nil, false
}

// getSubStates gets the substates of a state.
func (sg *StateGraph) getSubStates(state *State) []*State {
	var substates []*State
	if state.StateInfo != nil {
		for _, subInfo := range state.StateInfo.Substates {
			subName := fmt.Sprintf("%v", subInfo.UnderlyingState)
			if sub, exists := sg.States[subName]; exists {
				substates = append(substates, sub)
			}
		}
	}
	return substates
}

// isDecision checks if a state name refers to a decision node.
func (sg *StateGraph) isDecision(state *State) bool {
	for _, dec := range sg.Decisions {
		if dec.NodeName == state.NodeName {
			return true
		}
	}
	return false
}
