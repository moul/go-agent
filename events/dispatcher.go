// Package events contains an event dispatcher loosely inspired by the PSR-14
// specification, published under the MIT license by the PHP-FIG.
//
// See https://www.php-fig.org/psr/psr-14/ for details and background.
package events

import (
	"context"
	"fmt"
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

	// AddProviders sets the ListenerProviders for Events with a given Topic.
	// It returns the modified provider, making the call chainable.
	AddProviders(Topic, ...ListenerProvider) Dispatcher

	// Reset re-initializes the list of providers for the specified Topic values,
	// returning the dispatcher without any listener provider for those.
	Reset(topics ...Topic) Dispatcher
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

type providersMap map[Topic][]ListenerProvider

// dispatcher is the default implementation of the Dispatcher interface.
type dispatcher struct {
	m         sync.Mutex
	providers providersMap
}

func (d *dispatcher) Dispatch(ctx context.Context, e Event) (Event, error) {
	providers, ok := d.providers[e.Topic()]
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

	for _, provider := range providers {
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
	}
	return e, nil
}

func (d *dispatcher) AddProviders(topic Topic, providers ...ListenerProvider) Dispatcher {
	d.m.Lock()
	defer d.m.Unlock()
	if d.providers == nil {
		d.providers = make(providersMap)
	}
	d.providers[topic] = append(d.providers[topic], providers...)
	return d
}

// Reset is part of the Dispatcher interface.
func (d *dispatcher) Reset(topics ...Topic) Dispatcher {
	d.m.Lock()
	defer d.m.Unlock()
	// Shortcut to reset all topics, or uninitialized dispatcher instances.
	if d.providers == nil || len(topics) == 0 {
		d.providers = make(providersMap)
		return d
	}

	for _, topic := range topics {
		delete(d.providers, topic)
	}
	return d
}

// NewDispatcher returns a basic Dispatcher implementation.
//
// Client code may use this constructor or create their own Dispatcher implementations.
func NewDispatcher() Dispatcher {
	return &dispatcher{
		providers: make(providersMap),
	}
}

