package fsm

import (
	"context"
	"reflect"
	"testing"
)

type TestStruct struct {
	State State
}

func IsTestStructValid(ctx context.Context, e *Event) (bool, error) {
	return true, nil
}

func IsTestStructInvalid(ctx context.Context, e *Event) (bool, error) {
	return false, nil
}

func TestSetState(t *testing.T) {
	testStruct := &TestStruct{
		State: State("started"),
	}

	fsm := NewFSM()

	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name:   "make",
		From:   []State{"started"},
		To:     State("finished"),
		Guards: []Guard{IsTestStructValid},
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	err := fsm.Fire(context.Background(), testStruct, "make")

	if err != nil {
		t.Errorf("error = %v", err)
	}

	if testStruct.State != State("finished") {
		t.Error("expected state to be 'finished'")
	}
}

func TestInvalidTransition(t *testing.T) {
	testStruct := &TestStruct{
		State: State("started"),
	}

	fsm := NewFSM()

	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name:   "make",
		From:   []State{"started"},
		To:     State("finished"),
		Guards: []Guard{IsTestStructInvalid},
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	err := fsm.Fire(context.Background(), testStruct, "make")

	if e, ok := err.(InvalidTransitionError); !ok && e.Event != "make" && e.State != "started" {
		t.Error("expected 'InvalidTransitionError'")
	}
}

func TestInvalidEvent(t *testing.T) {
	testStruct := &TestStruct{
		State: State("started"),
	}

	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name:   "make",
		From:   []State{"started"},
		To:     State("finished"),
		Guards: []Guard{IsTestStructInvalid},
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	err := fsm.Fire(context.Background(), testStruct, "some_event_name")

	if e, ok := err.(UnknownEventError); !ok && e.Event != "some_event_name" {
		t.Error("expected 'UnknownEventError'")
	}
}

func TestPermittedEvents(t *testing.T) {
	testStruct := &TestStruct{
		State: State("started"),
	}

	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{"started"},
		To:   State("finished"),
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	permittedEvents, err := fsm.GetPermittedEvents(context.Background(), testStruct)
	if err != nil {
		t.Errorf("fsm.GetPermittedEvents() error = %v", err)
	}

	if len(permittedEvents) == 0 {
		t.Error("expected permitted events to be ['make']")
	}
}

func TestUnknownSrcState(t *testing.T) {
	testStruct := &TestStruct{
		State: State("started"),
	}

	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{"finished"},
		To:   State("started"),
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	permittedEvents, err := fsm.GetPermittedEvents(context.Background(), testStruct)
	if err != nil {
		t.Errorf("fsm.GetPermittedEvents() error = %v", err)
	}

	if len(permittedEvents) != 0 {
		t.Error("expected len permitted events to be 0")
	}
}

func TestPermittedEventsSkipGuards(t *testing.T) {
	testStruct := &TestStruct{
		State: State("started"),
	}

	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{"started"},
		To:   State("finished"),
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	permittedEvents, err := fsm.GetPermittedEvents(context.Background(), testStruct, SkipGuard(true))
	if err != nil {
		t.Errorf("fsm.GetPermittedEvents() error = %v", err)
	}

	if len(permittedEvents) == 0 {
		t.Error("expected permitted events to be ['make']")
	}
}

func TestPermittedStates(t *testing.T) {
	startedState := State("started")
	finishedState := State("finished")

	testStruct := &TestStruct{
		State: startedState,
	}

	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{startedState},
		To:   finishedState,
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	permittedStates, err := fsm.GetPermittedStates(context.Background(), testStruct)
	if err != nil {
		t.Errorf("fsm.GetPermittedStates() error = %v", err)
	}

	if len(permittedStates) == 0 {
		t.Errorf("expected permitted state to be %v", finishedState)
	}
}
