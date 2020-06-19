// Package events contains an event dispatcher loosely inspired by the PSR-14
// specification, published under the MIT license by the PHP-FIG.
//
// See https://www.php-fig.org/psr/psr-14/ for details and background.
package events

import (
	"net/http"
	"regexp"
)

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

	Request() *http.Request
	Response() *http.Response
	Err() error

	// SetData is a setter for the data returned by Data.
	SetData(interface{}) Event
	// SetRequest is a setter for the event request, returning the event.
	SetRequest(r *http.Request) Event
	// SetResponse is a setter for the event response, returning the event.
	SetResponse(r *http.Response) Event
	// SetError is a setter for the error, returning the event.
	SetError(err error) Event
}

// EventBase is a basic event implementation, meant to be composed into actual
// event types to provide default storage and code for the Event methods.
type EventBase struct {
	data     interface{}
	topic    Topic
	request  *http.Request
	response *http.Response
	Error    error
}

// Data is part of the Event interface.
func (eb *EventBase) Data() interface{} {
	return eb.data
}

// Request returns the http.Request in the event, which may be nil.
func (eb *EventBase) Request() *http.Request {
	return eb.request
}

// SetRequest sets the http.Request in the event, which may be nil.
func (eb *EventBase) SetRequest(r *http.Request) Event {
	eb.request = r
	return eb
}

// Response returns the http.Response in the event, which may be nil.
func (eb *EventBase) Response() *http.Response {
	return eb.response
}

// SetResponse sets the http.Response in the event, which may be nil.
func (eb *EventBase) SetResponse(r *http.Response) Event {
	eb.response = r
	return eb
}

// Err returns the error present on the event.
func (eb *EventBase) Err() error {
	return eb.Error
}

// SetError sets the error in the event, which may be nil.
func (eb *EventBase) SetError(err error) Event {
	eb.Error = err
	return eb
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

// NewEvent returns a basic Event implementation.
//
// It takes a string argument instead of a Topic to avoid unwary users passing
// an incorrect topic by just doing NewEvent(Topic(somestring)), thus bypassing
// topic format checks.
func NewEvent(topic string) Event {
	e := EventBase{}.WithTopic(topic)
	return e
}
