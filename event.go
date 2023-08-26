package turtle

import (
	"fmt"
	"reflect"
	"time"
)

// Event represents an event.
type Event struct {
	// Name - event name
	Name string
	// At - event time
	At time.Time
	// Attrs - optional event attributes
	Attrs map[string]any
}

const (
	eventAttrWorkerId = "workerId"
	eventAttrError = "error"
)

// EventSettings configures an event.
type EventSettings func(e *Event)

// WithEventWorkerId binds the worker id of the event attrs.
func WithEventWorkerId(workerId int) EventSettings {
	return func(e *Event) {
		if e.Attrs == nil {
			e.Attrs = make(map[string]any)
		}
		e.Attrs[eventAttrWorkerId] = workerId
	}
}

// WithEventError binds the error of the event attrs.
func WithEventError(err error) EventSettings {
	return func(e *Event) {
		if e.Attrs == nil {
			e.Attrs = make(map[string]any)
		}
		e.Attrs[eventAttrError] = err
	}
}

// NewEvent creates a new event.
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

// EventHandler handles events.
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

// NilEventHandler is an event handler that does nothing.
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