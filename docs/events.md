# Events system

To enable a good level of flexibility in the Agent and provide a non-intrusive
way for clients to interact with it if needed, the Agent includes a basic event
mechanisms.

It triggers events in a number of situations, like initialization, termination, 
configuration updates, or request handling steps.

It is also open for use by client applications for other purposes.


## Design

The event system provides :

- subscription to a topic by name and weight
- dispatching any kind of data to subscribers, with synchronous flow and defined order (by weight), accumulating event
  changes by all subscribers
- ability for synchronous subscribers to terminate dispatching
- support for "all" execution of subscribers implementing asynchronous execution: all subscribers can run concurrently,
  but dispatching is only done once all asynchronous subscribers finish.

Since it is an _ad hoc_ mechanism, it does not include features present in
fully-featured event systems, like:

- unsubscribe mechanism
- "once" subscriptions
- direct asynchronous subscriptions: async subscriptions are sync subscriptions launching a goroutine on their own

Inspirations:

- Asaskevich/EventBus (MIT)
- Doctrine EventManager (MIT)
- Symfony EventDispatcher (MIT)
- Node.JS EventEmitter (MIT)
