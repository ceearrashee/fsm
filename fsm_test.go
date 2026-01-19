package fsm

import (
	"context"
	"reflect"
	"sync"
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

func TestConcurrentFireOnDifferentInstances(t *testing.T) {
	// Test that Fire can be called concurrently on different instances without blocking
	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{"started"},
		To:   State("finished"),
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	// Create multiple instances
	instances := make([]*TestStruct, 10)
	for i := 0; i < 10; i++ {
		instances[i] = &TestStruct{State: State("started")}
	}

	// Fire events concurrently on all instances
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := fsm.Fire(context.Background(), instances[idx], "make")
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Fire() error = %v", err)
	}

	// Verify all instances transitioned
	for i, inst := range instances {
		if inst.State != State("finished") {
			t.Errorf("instance %d: expected state 'finished', got '%s'", i, inst.State)
		}
	}
}

func TestFireWithDependentObjectInCallback(t *testing.T) {
	// Test that calling Fire on a dependent object within a callback doesn't cause deadlock
	fsm := NewFSM()

	// Create two related instances
	instance1 := &TestStruct{State: State("started")}
	instance2 := &TestStruct{State: State("started")}

	// Register FSM with callback that fires event on dependent object
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{"started"},
		To:   State("finished"),
		After: func(ctx context.Context, e *Event) error {
			// This callback fires an event on a different instance
			// Previously this would cause a deadlock with global lock
			if e.Source == instance1 {
				return fsm.Fire(ctx, instance2, "make")
			}
			return nil
		},
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	// Fire on instance1, which will trigger Fire on instance2 in callback
	err := fsm.Fire(context.Background(), instance1, "make")
	if err != nil {
		t.Errorf("Fire() error = %v", err)
	}

	// Verify both instances transitioned
	if instance1.State != State("finished") {
		t.Errorf("instance1: expected state 'finished', got '%s'", instance1.State)
	}
	if instance2.State != State("finished") {
		t.Errorf("instance2: expected state 'finished', got '%s'", instance2.State)
	}
}

func TestReleaseInstance(t *testing.T) {
	// Test that Release properly cleans up instance locks
	fsm := NewFSM()
	if err := fsm.Register(reflect.TypeOf((*TestStruct)(nil)), "State", Events{{
		Name: "make",
		From: []State{"started"},
		To:   State("finished"),
	}}); err != nil {
		t.Errorf("fsm.Register() error = %v", err)
	}

	instance := &TestStruct{State: State("started")}

	// Fire an event to create a lock for this instance
	err := fsm.Fire(context.Background(), instance, "make")
	if err != nil {
		t.Errorf("Fire() error = %v", err)
	}

	// Release the instance
	fsm.Release(instance)

	// Fire again - should work and create a new lock
	instance.State = State("started")
	err = fsm.Fire(context.Background(), instance, "make")
	if err != nil {
		t.Errorf("Fire() after Release error = %v", err)
	}

	if instance.State != State("finished") {
		t.Errorf("expected state 'finished', got '%s'", instance.State)
	}
}
