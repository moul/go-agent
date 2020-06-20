# Bearer.sh Go agent

This module provides a pure Go HTTP (HTTPS, HTTP/2) transport decorator for Go
Web API clients. It relies on the https://bearer.sh platform to provide
metrics observation and anomaly detection.


## Getting started

Register on https://login.bearer.sh/login to obtain an account and its secret key.

Then, with that key in hand, just start your program code with a single specific statement:

```go
package main

import bearer `github.com/bearer/go-agent`

func main() {
    // Don't forget the second set of parentheses.
    defer bearer.Init(`app_50-digits-long-secret-key-from-bearer.sh`)()

    // From then on, all your http.Get and similar calls are instrumented.
}
```

This one-line version will even provide instrumentation for multiple clients using
the `http.DefaultTransport` provided by the Go standard library.


## Fully configurable version

If your needs go beyond the default Go transport, you can create your own instrumented
transport, and use it in your custom-created HTTP clients:

```go
package main 
import (
	"net/http"
	"os"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/proxy"
)


func main() {
	// Step 1: initialize Bearer.
	defer agent.Init(os.Getenv(config.SecretKeyName))()

	// Step 2: prepare your custom transport, decorating it with the Bearer agent.
	var baseTransport = http.DefaultTransport.(*http.Transport)

	// Say your enterprise proxy needs a specific CONNECT header.
	baseTransport.ProxyConnectHeader = http.Header{"ACME_ID": []string{"some secret"}}
	transport := agent.DefaultAgent.Decorate(baseTransport)

	// Step 3: use your client as usual.
	client := http.Client{Transport: transport}
	response, err := client.Do(&http.Request{URL: proxy.MustParseURL(`http://someurl.tld/path`)})

    // ...use the API response and error as usual
}
```

For even more advanced use cases, the API also allows creating multiple
configurations and multiple Bearer agents.


## Privacy considerations  / GDPR

Since logging API calls may involve sensitive data, you may want to configure
custom filters to strip [PII](https://gdpr.eu/eu-gdpr-personal-data/) from the 
logs, using the `SensitiveKeys` and `SensitiveRegex` options on the agent.


## Deployment

On a live system, you will likely apply two best practices:

- Take the secret key from the environment. We suggest calling the variable
  `BEARER_SECRETKEY`, for which the `SecretKeyName` constant is available in the
  `config` package.
- For logging
  - either use the default agent logging, which goes to standard error output 
    (12-factor suggests standard output), whence messages can be picked up, 
  - or apply a frequent deviation from 12-factor by injecting a logger of your
    choice to the configuration in `config/NewConfig()` and, ideally, also
    injecting it to the default Go logger using `log.SetOutput(myLogger)` to
    ensure logs consistency.
     
Your firewall will need to allow your application to perform outgoing HTTPS/HTTP2
calls to the Bearer platform, at `https://config.bearer.sh` and `https://logs.bearer.sh`.


## Prerequisites

- For your applications:
  - Go 1.13 or later
  - Go modules enabled
- To contribute to the agent
  - Go [`stringer`](https://pkg.go.dev/golang.org/x/tools/cmd?tab=overview) command
  - The `make` command    
  - To rebuild the import graph
    - [Godepgraph](https://github.com/kisielk/godepgraph) command
    - [Graphviz](https://graphviz.org/)
  - To check syntax and coding standards:
    - [Golint](https://github.com/golang/lint)
    - [Golangci-lint](https://github.com/golangci/golangci-lint)
    

## Contributing

Please read [CONTRIBUTING.md](https://example.com) for details on our code of 
conduct, and the process for submitting pull requests to us.


### Running the tests

The run the 450+ tests in the package, you can use `go test` if you wish, or run
the preconfigured `go test` commands in the `Makefile`:

- `make test_quick` runs the tests as fast as possible, not checking for race conditions
- `make test_racy` runs the tests with the race detector, making them significantly slower


### Run coding style tests

These tests verify that the code base applies best practices: `make lint`

This should just show the command being run, and display no warnings.

### Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available,
see the [tags on this repository](https://code.osinet.fr/OSInet/bearer-go-agent/releases). 

- Versions 2.m.p are the current stable versions.
- Versions 0.m.p and 1.m.p use the original PoC code base and are now obsolete.

When preparing to commit a branch, commit all files, then, run `go generate agent.go` 
to generate the `agent_sha.go` file containing the commit SHA, and add a commit
with that information, not squashing it. Users will be able to report that SHA
to support to enable them to be sure of the version of the agent actually in use.
 

## Credits / Legal
### Authors

- **Frédéric G. MARAND** - *Project development* - [OSInet](https://osinet.fr/go)
- **Manfred TOURON** - *PoC version* - [Manfred.life](https://manfred.life)
- **Billie THOMPSON** - *"Contributing" template* - [PurpleBooth](https://github.com/PurpleBooth)

<!-- See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project. -->


### License

This project is published under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details


### Acknowledgments

- The events package is very much inspired by the 
  [PSR-14](https://www.php-fig.org/psr/psr-14/) specification, published under 
  the CC-BY-3.0 UNPORTED license for text and MIT License for code.
- The dependency resolution algorithm in `Description.resolveHashes` is adapted
  from an [article by Ferry Boender], published under a permissive license:
  
        This document may be freely distributed, in part or as a whole, on any
        medium, without the prior authorization of the author, provided that this
        Copyright notice remains intact, and there will be no obstruction as to
        the further distribution of this document. You may not ask a fee for the
        contents of this document, though a fee to compensate for the distribution
        of this document is permitted.
    
        Modifications to this document are permitted, provided that the modified
        document is distributed under the same license as the original document
        and no copyright notices are removed from this document. All contents
        written by an author stays copyrighted by that author.
    
        Failure to comply to one or all of the terms of this license automatically
        revokes your rights granted by this license
    
        All brand and product names mentioned in this document are trademarks or
        registered trademarks of their respective holders.
- The well-formed invalid credit card numbers used for sensitive data validation
  were provided by https://www.freeformatter.com/credit-card-number-generator-validator.html
  
[article by Ferry Boender]: https://www.electricmonk.nl/docs/dependency_resolving_algorithm/dependency_resolving_algorithm.html
