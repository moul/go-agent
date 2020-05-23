package events_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bearer/go-agent/events"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    events.Error
		want string
	}{
		{"Propagation stop request", events.DispatchStopRequest, string(events.DispatchStopRequest)},
		{"empty", events.Error(""), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventBase_Data(t *testing.T) {
	var nilSlice []byte
	var chanBool = make(chan bool, 5)

	tests := []struct {
		name string
		data interface{}
		want interface{}
	}{
		{"nil-nil", nil, nil},
		{"nil-slice", nilSlice, nilSlice},
		{"channel", chanBool, chanBool},
		{"string", "foo", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := &events.EventBase{}
			eb.SetData(tt.data)
			eb.WithTopic(tt.name)
			if got := eb.Data(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Data() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventBase_Topic(t *testing.T) {
	tests := []struct {
		name    string
		inbound string
		want    events.Topic
	}{
		{"happy", "happy", "happy"},
		{"replaced", "ill formed", "ill-formed"},
		{"empty", "", events.TopicEmpty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := events.NewEvent(tt.inbound)
			if got := eb.Topic(); got != tt.want {
				t.Errorf("Topic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListenerProviderFunc_getListenersForEvent(t *testing.T) {
	eb := events.EventBase{}
	eventDependent := func(e events.Event) []events.Listener {
		if e == nil {
			return nil
		}
		return []events.Listener{}
	}
	tests := []struct {
		name string
		lpf  events.ListenerProviderFunc
		want []events.Listener
	}{
		{"nil", nil, nil},
		{"nil-returning", func(events.Event) []events.Listener { return nil }, nil},
		{"happy", eventDependent, []events.Listener{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lpf.Listeners(&eb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Listeners() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_dispatcher_SetProvider asserts that the ListenerProvider assignment is
// correct by checking how the Listeners they return act upon dispatch.
func Test_dispatcher_SetProvider_bb(t *testing.T) {
	const (
		t1 = "topic-1"
		t2 = "topic-2"
	)
	var l1 events.Listener = func(_ context.Context, e events.Event) error {
		e.SetData(t1)
		return nil
	}
	var l2 events.Listener = func(_ context.Context, e events.Event) error {
		e.SetData(t2)
		return nil
	}
	var p1 events.ListenerProviderFunc = func(events.Event) []events.Listener {
		return []events.Listener{l1}
	}
	var p2 events.ListenerProviderFunc = func(events.Event) []events.Listener {
		return []events.Listener{l2}
	}
	e1 := events.NewEvent(t1)
	e2 := events.NewEvent(t2)
	bg := context.Background()
	d := events.NewDispatcher()

	d.SetProvider(t1, p1).SetProvider(t2, p2)
	tests := []struct {
		name string
		p    events.ListenerProvider
		e    events.Event
	}{
		{t1, p1, e1},
		{t2, p2, e2},
	}
	for _, tt := range tests {
		ev, err := d.Dispatch(bg, tt.e)
		if err != nil {
			t.Fatalf("unexpected error during dispatching: %v", err)
		}
		actual, ok := ev.Data().(string)
		if !ok {
			t.Fatalf("got unexpected data type after dispatch: %#v", ev.Data())
		}
		if actual != tt.name {
			t.Fatalf("expected %s, got %s", tt.name, actual)
		}
	}
}

func Test_dispatcher_DispatchWithoutProvider(t *testing.T) {
	const topic = "topic"
	const data = 42
	ctx := context.Background()
	d := events.NewDispatcher()
	e := events.NewEvent(topic).SetData(data)
	e, err := d.Dispatch(ctx, e)
	if err != nil {
		t.Fatalf("failed dispatching without any provider: %v", err)
	}
	if e == nil {
		t.Fatalf("dispatching returned nil event without an error")
	}
	if actual := e.Topic(); actual != topic {
		t.Fatalf("topic changed without any provider, expected %s, got %s", actual, e.Topic())
	}
	actual, ok := e.Data().(int)
	if !ok {
		t.Fatalf("Data type changed without any provider. Expected int, got %T", actual)
	}
	if actual != data {
		t.Fatalf("Data value changed without any provider. Expected %d, got %d", data, actual)
	}
}

func Test_dispatcher_DispatchCanceledContext(t *testing.T) {
	const topic = "topic"
	const data = 42
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	lp := events.ListenerProviderFunc(func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				cancel()
				return nil
			},
		}
	})

	d := events.NewDispatcher().SetProvider(topic, lp)
	e := events.NewEvent(topic).SetData(data)
	e, err := d.Dispatch(ctx, e)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("failed to abort dispatch on timed out context: %v", err)
	}
	if err == nil {
		t.Fatalf("failed reporting error on timed out context: %v", err)
	}
	if e == nil {
		t.Fatalf("dispatching returned nil event without an error")
	}
	if actual := e.Topic(); actual != topic {
		t.Fatalf("topic changed with timed out context, expected %s, got %s", actual, e.Topic())
	}
	actual, ok := e.Data().(int)
	if !ok {
		t.Fatalf("Data type changed timed out context. Expected int, got %T", actual)
	}
	if actual != data {
		t.Fatalf("Data value changed timed out context. Expected %d, got %d", data, actual)
	}
}

func Test_dispatcher_DispatchStop(t *testing.T) {
	const topic = "topic"
	var lp events.ListenerProviderFunc

	d := events.NewDispatcher()
	e := events.NewEvent(topic)

	ctx := context.Background()
	lp = func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				return events.DispatchStopRequest
			},
		}
	}

	d.SetProvider(topic, lp)
	_, err := d.Dispatch(ctx, e)
	if err != nil {
		t.Fatalf("returned a non-nil error on stop request: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	lp = func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				cancel()
				return events.DispatchStopRequest
			},
		}
	}
	d.SetProvider(topic, lp)
	_, err = d.Dispatch(ctx, e)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("returned did not return a canceled error on stopping listener triggering cancellation: %v", err)
	}

}

func Test_dispatcher_DispatchCancel(t *testing.T) {
	const topic = "topic"
	ctx, cancel := context.WithCancel(context.Background())
	lp := events.ListenerProviderFunc(func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				cancel()
				return nil
			},
		}
	})
	d := events.NewDispatcher().SetProvider(topic, lp)
	_, err := d.Dispatch(ctx, events.NewEvent(topic))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("returned a non-Canceled context error: %v", err)
	}
}

func Test_dispatcher_DispatchTimeout(t *testing.T) {
	const topic = "topic"
	const delay = 10*time.Microsecond
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	defer cancel()

	lp := events.ListenerProviderFunc(func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				// Be sure to exceed timeout.
				time.Sleep(2*delay)
				return nil
			},
		}
	})
	d := events.NewDispatcher().SetProvider(topic, lp)
	_, err := d.Dispatch(ctx, events.NewEvent(topic))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("returned a non-DeadlineExceeded context error: %v", err)
	}
}

func Test_dispatcher_DispatchError(t *testing.T) {
	const topic = "topic"
	const expected = events.Error("random error")

	ctx := context.Background()
	lp := events.ListenerProviderFunc(func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				return expected
			},
		}
	})
	d := events.NewDispatcher().SetProvider(topic, lp)
	_, err := d.Dispatch(ctx, events.NewEvent(topic))
	if !errors.Is(err, expected) {
		t.Fatalf("returned an unexpected error: %v", err)
	}
}

func Test_dispatcher_DispatchCancelAndError(t *testing.T) {
	const topic = "topic"
	const expected = events.Error("random error")

	ctx, cancel := context.WithCancel(context.Background())
	lp := events.ListenerProviderFunc(func(events.Event) []events.Listener {
		return []events.Listener{
			func(context.Context, events.Event) error {
				cancel()
				return expected
			},
		}
	})
	d := events.NewDispatcher().SetProvider(topic, lp)
	_, err := d.Dispatch(ctx, events.NewEvent(topic))
	if !errors.Is(err, expected) {
		t.Fatalf("returned an unexpected listener error: %v", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, "context canceled") {
		t.Errorf("returned a non-Canceled context error: %v", err)
	}
}

