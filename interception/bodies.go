package interception

import (
	"context"

	"github.com/bearer/go-agent/events"
)

type BodyParsingProvider struct {
}

func (p BodyParsingProvider) Listeners(e events.Event) []events.Listener {
	if e.Topic() != TopicBodies {
		return nil
	}

	return []events.Listener{
		p.ParseRequestBody,
		p.ParseResponseBody,
	}
}

func (p BodyParsingProvider) ParseRequestBody(_ context.Context, e events.Event) error {

	return nil
}

func (p BodyParsingProvider) ParseResponseBody(_ context.Context, e events.Event) error {
	return nil
}
