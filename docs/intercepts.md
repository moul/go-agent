# Intercepts
## Simple cases

For simple case, Go clients tend to use the `http.DefaultClient`. This means
that the Bearer agent can swap it with an instrumented version and simple calls
will just work with no extra work, like the following:

```go
	// Step 1: initialize the Bearer agent.
	defer agent.Init(secretKey)()

	// Step 2: use your API as usual.
	res, _ := http.Get(exampleAPIURL.String())
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	// Step 3: there is no need for a third step !
}
```

See a more complete simple example in `examples/simple/simple.go`.

If your code uses multiple custom clients with various settings, but still uses
the `http.DefaultTransport`, they will also be instrumented with no extra work.


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
	defer foo.Body.Close()
	fooBody := ioutil.ReadAll(foo.Body)
``` 


## Advanced use

For advanced use cases, you can create multiple agents, and each of them will
maintain an isolated Bearer profile. Following the example of the Go runtime library,
the `Init` function is actually just a helper using a default agent provided by
the package, covering the 99% use cases where the application only uses a single
Bearer agent.

```go
	// Step 1: initialize the Bearer agents.
	agent1 := agent.NewAgent()
    agent2 := agent.NewAgent()
    
    defer agent1.Init(secretKey1)() 
    defer agent2.Init(secretKey2)()

	// Step 2: prepare your custom transport, decorating it with the Bearer agent.
	var fooTransport http.RoundTripper = agent1.Decorate(NewFooTransport())
	var barTransport http.RoundTripper = agent2.Decorate(NewBarTransport())

	// Step 3: use your client as usual.
	fooClient := http.Client{ Transport: fooTransport }
    barClient := http.Client{ Transport: barTransport }
        
	foo, _ := fooClient.Do(&http.Request{URL: someURL)
	bar, _ := barClient.Do(&http.Request{URL: someOtherURL)
	defer foo.Body.Close()
    defer bar.Body.Close()
	fooBody := ioutil.ReadAll(foo.Body)
    barBody := ioutil.ReadAll(bar.Body))
``` 
 
                                     

