# Intercepts
## Simple cases

For simple case, Go clients tend to use the `http.DefaultClient`. This means
that the Bearer agent can swap it with an instrumented version and simple calls
will just work with no extra work, like the following:

```go
	// Step 1: initialize the Bearer agent.
	defer agent.Init(secretKey)()

	// Step 2: use your API as usual.
	res, _ := http.DefaultClient.Get(exampleAPIURL.String())
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	// Step 3: there is no need for a third step !
}
```

See a more complete simple example in `examples/simple/simple.go`.

If your code uses multiple custom clients with various settings, but still uses
the `http.DefaultTransport`, they will also be instrumented with no extra work.

You may also pass extra clients to the `agent.Init` call, if they are already
available on application startup, and they will all be monitored, like this:

```go
	// Step 1: initialize the Bearer agent, adding the custom client(s).
	defer agent.Init(secretKey, fooClient)()

	// Step 2: use your custom client as usual.
	res, _ := fooClient.Get(exampleAPIURL.String())
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
```


## Bespoke clients

For bespoke code, using the Go runtime in more sophisticated ways, the default
Go transport is sometimes not sufficient.

In that case, you will need to wrap your custom transports with the Bearer agent,
as in the following example:

```go
	// Step 1: initialize the Bearer agent.
	defer agent.Init(secretKey)()

	// Step 2: prepare your custom transport, decorating it with the Bearer agent.
	var fooTransport http.RoundTripper = agent.Decorate(NewFooTransport())

	// Step 3: use your client as usual.
	fooClient := http.Client{ Transport: fooTransport }
	foo, _ := fooClient.Do(&http.Request{URL: someURL)
	defer bar.Body.Close()
	fooBody := ioutil.Readall(bar.Body)
``` 
