package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bearer/go-agent/examples"

	bearer "github.com/bearer/go-agent"
)

func main() {
	secretKey := os.Getenv(bearer.SecretKeyName)
	if len(secretKey) == 0 {
		log.Fatalf(`Bearer needs a %s environment variable`, bearer.SecretKeyName)
	}

	// Step 1: initialize Bearer.
	//
	// Creating an agent sets up Bearer decoration for the DefaultClient,
	// allowing any API call using it to be monitored.
	//
	// The Close method can be used to cleanly shut down the agent and wait for
	// any outstanding API calls to be reported
	//
	// Note that, since the Go runtime httptest uses manually defined clients,
	// your running HTTP tests will not trigger extra monitoring calls to Bearer.
	agent := bearer.New(secretKey)
	defer agent.Close()

	// Step 2: use the default Go client as usual to perform your API call.
	//
	// The client will trigger monitoring for the request parameters, and the
	// request and response headers.
	examples.APIURL = `https://code.osinet.fr/api/v1/orgs/OSInet?a=a11&a=a12&b=2&password=secret&foo=her+email+is+jane.doe@example.com&card=4539530418912307`
	for i := 0; i < 10; i++ {
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
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatalf("reading API response: %v", err)
		}

		examples.ShowGogsOrg(body)
		time.Sleep(700 * time.Millisecond)
	}
}
