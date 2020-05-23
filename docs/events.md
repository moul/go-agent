# Events system

To enable a good level of flexibility in the Agent and provide a non-intrusive
way for clients to interact with it, if needed, the Agent includes a basic event
mechanisms.

It triggers events in a number of situations, like initialization, termination, 
configuration updates, or request handling steps.

It is also open for use by client applications for other purposes.


## Design

The event system provides :

- subscription to a topic by listener provides
- dispatching any kind of data to subscribers, with synchronous flow and defined
  order, accumulating event changes by all subscribers
- ability for synchronous subscribers to terminate dispatching

Since it is an _ad hoc_ mechanism, it does not include features present in
fully-featured event systems, like:

- support for "all" execution of subscribers implementing asynchronous execution:
  all subscribers can run concurrently, but dispatching would be done only done
  once all asynchronous subscribers finish.
- unsubscribe mechanism
- "once" subscriptions
- direct asynchronous subscriptions: async subscriptions are sync subscriptions launching a goroutine on their own


Inspirations:

- The PSR-14 specification (MIT) is the general inspiration for the design,
  especially for the listener providers concept.
- Asaskevich/EventBus (MIT)
- Doctrine EventManager (MIT)
- Symfony EventDispatcher (MIT)
- Node.JS EventEmitter (MIT)
