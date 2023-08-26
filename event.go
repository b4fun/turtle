package turtle

import (
	"fmt"
	"reflect"
	"time"
)

type Event struct {
	Name string
	At time.Time
	Attrs map[string]any
}

const (
	eventAttrWorkerId = "workerId"
	eventAttrError = "error"
)

type EventSettings func(e *Event)

func WithEventWorkerId(workerId int) EventSettings {
	return func(e *Event) {
		if e.Attrs == nil {
			e.Attrs = make(map[string]any)
		}
		e.Attrs[eventAttrWorkerId] = workerId
	}
}

func WithEventError(err error) EventSettings {
	return func(e *Event) {
		if e.Attrs == nil {
			e.Attrs = make(map[string]any)
		}
		e.Attrs[eventAttrError] = err
	}
}

func NewEvent(eventName string, settings ...EventSettings) Event {
	rv := Event{
		Name: eventName,
		At: time.Now(),
	}
	for _, s := range settings {
		s(&rv)
	}

	return rv
}

type EventHandler interface {
	HandleEvent(event Event)
}

type EventHandleFunc func(event Event)

func (f EventHandleFunc) HandleEvent(event Event) {
	f(event)
}

type curriedEventHandler struct {
	EventHandler
	settings []EventSettings
}

func (f *curriedEventHandler) HandleEvent(event Event) {
	e := event
	for _, s := range f.settings {
		s(&e)
	}
	f.EventHandler.HandleEvent(e)
}

func wrapEventSettings(f EventHandler, settings ...EventSettings) EventHandler {
	return &curriedEventHandler{
		EventHandler: f,
		settings: settings,
	}
}

func capturePanic(f EventHandler) func() {
	return func() {
		if r:= recover(); r != nil {
			err := fmt.Errorf("panic: %s", r)
			f.HandleEvent(NewEvent(EventWorkerPanic, WithEventError(err)))
		}
	}
}

var NilEventHandler EventHandler = EventHandleFunc(func(event Event) {})

func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

const (
	EventTCPDial = "tcp/dial"
	EventTCPClosed = "tcp/closed"

	EventWorkerError = "worker/error"
	EventWorkerPanic = "worker/panic"
)