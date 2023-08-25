package turtle

import (
	"reflect"
	"time"
)

type Event struct {
	Name string
	At time.Time
	// TODO: attrs?
}

type EventHandler interface {
	HandleEvent(event Event)
}

type EventHandleFunc func(event Event)

func (f EventHandleFunc) HandleEvent(event Event) {
	f(event)
}

var nilEventHandler EventHandler = EventHandleFunc(func(event Event) {})

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
)