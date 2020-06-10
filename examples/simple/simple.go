package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/config"
	"github.com/bearer/go-agent/examples"
)

func main() {
	// Step 1: initialize Bearer.
	//
	// agent.Init(secretKey) returns a closer function, which never fails.
	// defer-ing it ensures it will run in all normal function return cases, as
	// well as panics. It will only fail to return if os.Exit() is called, as
	// that function exits the program without calling any deferred code.
	//
	// This single step sets up Bearer decoration for the DefaultClient, allowing
	// any API call using it to be monitored.
	//
	// Note that, since the Go runtime httptest uses manually defined clients,
	// your running HTTP tests will not trigger extra monitoring calls to Bearer.
	secretKey := os.Getenv(config.SecretKeyName)
	if !config.SecretKeyRegex.MatchString(secretKey) {
		secretKey = agent.ExampleWellFormedInvalidKey
	}
	defer agent.Init(secretKey)()

	// Step 2: use the default Go client as usual to perform your API call.
	//
	// The client will trigger monitoring for the request parameters, and the
	// request and response headers.
	examples.APIURL = `https://code.osinet.fr/api/v1/orgs/OSInet?a=a11&a=a12&b=2&password=secret&foo=her+email+is+jane.doe@example.com&card=4539530418912307`
	for i := 0; i < 1; i++ {
		res, err := http.Get(examples.APIURL)
		if err != nil {
			log.Fatalf("calling %s: %v", examples.APIURL, err)
		}

		// Step 3: use the standard response body as usual.
		//
		// The response Body is decorated by a monitoring mechanism tracking the API
		// request and response bodies: since go supports fully multiplexed HTTP and
		// HTTP/2, the request body may not be entirely available when the request
		// starts, but only when the response ends.
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("reading API response: %v", err)
		}

		examples.ShowGogsOrg(body)
		time.Sleep(700 * time.Millisecond)
	}
}
