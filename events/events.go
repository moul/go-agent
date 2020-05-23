// Package events contains an event dispatcher loosely inspired by the PSR-14
// specification, published under the MIT license by the PHP-FIG.
//
// See https://www.php-fig.org/psr/psr-14/ for details and background.
package events

import (
	"context"
	"fmt"
	"regexp"
	"sync"
)

// Error provides the ability to define constant errors, preventing global modification.
// It is based on the https://dave.cheney.net/2016/04/07/constant-errors article
// by Dave Cheney (CC-4.0-BY-NC-SA).
type Error string

// Error implements the error interface.
func (e Error) Error() string {
	return string(e)
}

// DispatchStopRequest is a sentinel error used by listeners to request
// that propagation be stopped, without any error condition being reported to
// the Emitter.
const DispatchStopRequest = Error("stop dispatch")

// TopicFormat is the format of strings used as Event Topics.
const TopicFormat = `^[-_[:alnum:]]+$`

// TopicReplacement is the format used to replace non-well-formed Topic strings.
const TopicReplacement = `[^_[:alnum:]]+`

// TopicEmpty is the replacement string for empty topics
const TopicEmpty = "-empty-"

// Topic is the type used for event labeling.
//
// Unlike vanilla strings, Topic instances should match the TopicFormat regexp,
// for debugging convenience.
type Topic string

// Event is the type of data passed around by a Dispatcher to Listeners.
type Event interface {
	// Topic returns an administrative label for the event.
	Topic() Topic

	// Data returns a generic payload.
	//
	// Event implementations should provide more strictly typed interfaces for
	// Listeners to use, allowing them to assert for the actual expected type
	// and use the typed getters accordingly.
	Data() interface{}

	// SetData is a setter for the data returned by Data.
	SetData(interface{}) Event
}

// Dispatcher is the interface for event dispatchers. It is inspired by the
// PSR-14 EventDispatcherInterface.
//
// It is responsible for retrieving Listeners from a ListenerProvider for the
// Event to dispatch, and invoking each Listener with that Event.
//
// It calls Listeners synchronously in the order they are returned by the
// ListenerProvider, then return to the Emitter:
//
//   - after all Listeners have executed,
//   - or an error occurred,
//   - or listener requested propagation to stop,
//   - or the context was canceled.
type Dispatcher interface {
	// Dispatch provides all relevant listeners with an event to process.
	// It returns the same Event object it was passed after it is done invoking
	// Listeners. The Event may have been mutated by Listeners.
	// If an error or propagation stop was returned by one of the Listeners,
	// or if the context is canceled, the dispatch loop is terminated and
	// Dispatch returns that error, possibly wrapped with context data.
	Dispatch(context.Context, Event) (Event, error)

	// SetProvider sets the ListenerProvider for Events with a given Topic.
	// It returns the modified provider, making the call chainable.
	SetProvider(Topic, ListenerProvider) Dispatcher
}

// Listener is the type passed to Dispatchers as callbacks acting on events.
//
// Unlike PSR-14 listeners, they return an error which, if non-nil, stops
// propagation and will be returned to the Emitter.
//
// That error can be sentinel value DispatchStopRequest.
type Listener func(context.Context, Event) error

// ListenerProvider provides the list of Listeners a Dispatcher must invoke for
// a given event.
type ListenerProvider interface {
	Listeners(Event) []Listener
}

// ListenerProviderFunc provides a means for a plain function to implement the
// ListenerProvider interface.
type ListenerProviderFunc func(e Event) []Listener

// Listeners implements the ListenerProvider interface.
func (lpf ListenerProviderFunc) Listeners(e Event) []Listener {
	if lpf == nil {
		return nil
	}
	return lpf(e)
}

// EventBase is a basic event implementation, meant to be composed into actual
// event types to provide default storage and code for the Event methods.
type EventBase struct {
	data  interface{}
	topic Topic
}

// Data is part of the Event interface.
func (eb *EventBase) Data() interface{} {
	return eb.data
}

// SetData is part of the Event interface.
func (eb *EventBase) SetData(data interface{}) Event {
	eb.data = data
	return eb
}

// Topic implements the Event interface.
func (eb *EventBase) Topic() Topic {
	return eb.topic
}

// WithTopic returns a new event based on the original one, with the name
// changed, allowing the original event to remain un-mutated.
// If the requested topic does not match the expected Topic format, it is modified
// to match it.
func (eb EventBase) WithTopic(topic string) Event {
	reMatcher := regexp.MustCompile(TopicFormat)
	if !reMatcher.MatchString(topic) {
		reReplacer := regexp.MustCompile(TopicReplacement)
		topic = reReplacer.ReplaceAllLiteralString(topic, "-")
	}
	if topic == "" {
		topic = TopicEmpty
	}
	eb.topic = Topic(topic)
	return &eb
}

// dispatcher is the default implementation of the Dispatcher interface.
type dispatcher struct {
	m         sync.Mutex
	providers map[Topic]ListenerProvider
}

func (d *dispatcher) Dispatch(ctx context.Context, e Event) (Event, error) {
	provider, ok := d.providers[e.Topic()]
	// Shortcut: no provider means no listeners, so nothing to call.
	if !ok {
		return e, nil
	}

	contextualize := func(step int, stage string, err error) error {
		switch err {
		case context.Canceled:
			return fmt.Errorf("cancelled %s listener #%d: %w", stage, step, err)
		case context.DeadlineExceeded:
			return fmt.Errorf("deadline exceeded %s listener #%d: %w", stage, step, err)
		}
		return err
	}

	// Ensure any context-aware async code run by a listener is able to be canceled
	// when the dispatch loop ends.
	dispatcherCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, listener := range provider.Listeners(e) {
		var ctxErr error
		if ctxErr = dispatcherCtx.Err(); ctxErr != nil {
			return e, contextualize(i, "before", ctxErr)
		}
		listenerErr := listener(dispatcherCtx, e)
		if ctxErr = dispatcherCtx.Err(); ctxErr != nil {
			ctxErr = contextualize(i, "after", ctxErr)
		}

		switch listenerErr {
		case nil:
			if ctxErr != nil {
				return e, ctxErr
			}
			continue

		case DispatchStopRequest:
			if ctxErr != nil {
				return e, ctxErr
			}
			return e, nil

		default:
			if ctxErr == nil {
				return e, listenerErr
			}
			wle := fmt.Errorf("listener %d error: %w", i, listenerErr)
			wce := contextualize(i, "during", ctxErr)
			return e, fmt.Errorf("%w and %v", wle, wce)

		}
	}
	return e, nil
}

func (d *dispatcher) SetProvider(topic Topic, provider ListenerProvider) Dispatcher {
	d.m.Lock()
	defer d.m.Unlock()
	if d.providers == nil {
		d.providers = make(map[Topic]ListenerProvider, 1)
	}
	d.providers[topic] = provider
	return d
}

// NewDispatcher returns a basic Dispatcher implementation.
//
// Client code may use this constructor or create their own Dispatcher implementations.
func NewDispatcher() Dispatcher {
	return &dispatcher{
		providers: make(map[Topic]ListenerProvider, 1),
	}
}

// NewEvent returns a basic Event implementation.
func NewEvent(topic string) Event {
	e := EventBase{}.WithTopic(topic)
	return e
}
