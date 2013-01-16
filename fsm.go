// Package fsm implements simple Finite-State Machines.
package fsm

import "reflect"

// A Finite-State Machine
type FSM struct {
	self    interface{}
	current State
}

// Returns a new FSM with a given "self" value (see State) and an initial State.
func New(self interface{}, initial State) *FSM {
	initial.Enter(self)
	return &FSM{
		self:    self,
		current: initial,
	}
}

// Triggers an action on an FSM. Trigger will panic if the action is unknown or
// cannot be triggered from the current state. args will be passed to the
// current State's Trigger method.
func (f *FSM) Trigger(action string, args ...interface{}) (err error) {
	newstate, err := f.current.Trigger(f.self, action, args)
	if err == nil && newstate != nil {
		f.current.Exit(f.self)
		f.current = newstate
		f.current.Enter(f.self)
	}
	return
}

// State is responsible for handling triggered actions and state transitions.
type State interface {

	// Trigger is called when an action is triggered on an FSM.  The self
	// parameter will always be the value passed to New() when the FSM was
	// created.  If a non-nil State is returned, then the FSM will switch to that
	// state before returning.
	//
	// Note that if the current state is returned, then it's Exit and Enter
	// methods will still be called.  They will not be called if nil is returned.
	Trigger(self interface{}, action string, args []interface{}) (State, error)

	// Enter is called when a state becomes the current state, before any actions
	// are triggered.
	Enter(self interface{})

	// Exit is called when the FSM changes to another state.
	Exit(self interface{})
}

// Actions provides an easy way of specifying states.
//
// Actions maps action names (as they are passed to the FSM's Trigger method)
// to handler functions. Each handler function must accept an arbitrary value
// for "self" as the first parameter. If there are any extra input paramaters,
// then they must be passed when triggering the action, or else Trigger will
// panic. Each handler function may return no values, in which case Trigger
// will return nil (for both the State and error return values); one value,
// which must be of a type that is assignable to type State; or one value as
// described and an error value.
//
// Enter and Exit are implemented by using the special action names "_enter"
// and "_exit", respectively, if they exist in the map.
type Actions map[string]interface{}

// Triggers an action by calling the corresponding handler function. Trigger
// will panic if there is no entry for the given action.
func (a Actions) Trigger(self interface{}, action string, args []interface{}) (State, error) {
	handler := a.getActionHandler(action)
	if handler == nil {
		panic("action '" + action + "' cannot be triggered from the current state")
	}
	return handler(self, args)
}

// Enter will call the function associated with the "_enter" key, if it exists.
func (a Actions) Enter(self interface{}) {
	if handler := a.getActionHandler("_enter"); handler != nil {
		handler(self, nil)
	}
}

// Exit will call the function associated with the "_exit" key, if it exists.
func (a Actions) Exit(self interface{}) {
	if handler := a.getActionHandler("_exit"); handler != nil {
		handler(self, nil)
	}
}

func (a Actions) getActionHandler(action string) actionHandler {
	value := a[action]
	if value == nil {
		return nil
	}
	handler, ok := value.(actionHandler)
	if !ok {
		handler = newActionHandler(reflect.ValueOf(value))
		a[action] = handler
	}
	return handler
}

type actionHandler func(self interface{}, args []interface{}) (State, error)

func newActionHandler(handler reflect.Value) actionHandler {
	typ := handler.Type()

	// Check function signature
	if typ.Kind() != reflect.Func {
		panic("action handler must be a function")
	}
	if typ.NumIn() == 0 {
		panic("action handler must have at least one input parameter")
	}
	switch typ.NumOut() {
	case 2:
		if typ.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
			panic("second return value must be of type: error")
		}
		fallthrough
	case 1:
		if !typ.Out(0).AssignableTo(reflect.TypeOf((*State)(nil)).Elem()) {
			panic("first return value must be assignable to type: fsm.State")
		}
	case 0:
	default:
		panic("action handler must return at most two values")
	}

	return actionHandler(func(self interface{}, args []interface{}) (newstate State, err error) {
		params := make([]reflect.Value, 1+len(args))
		params[0] = reflect.ValueOf(self)
		for i, param := range args {
			params[i+1] = reflect.ValueOf(param)
		}
		results := handler.Call(params)
		if len(results) > 0 {
			reflect.ValueOf(&newstate).Elem().Set(results[0])
			if len(results) > 1 {
				reflect.ValueOf(&err).Elem().Set(results[1])
			}
		}
		return
	})
}
