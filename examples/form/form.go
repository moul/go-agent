package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/examples"
	"github.com/bearer/go-agent/proxy"
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
	secretKey := os.Getenv(agent.SecretKeyName)
	if len(secretKey) == 0 {
		log.Fatalf(`Bearer needs a %s environment variable`, agent.SecretKeyName)
	}
	defer agent.Init(secretKey)()

	// Step 2: use the default Go client as usual to perform your API call.
	//
	// The client will trigger monitoring for the request parameters, and the
	// request and response headers.
	examples.APIURL = `https://ac.audean.com/sc.php`
	res, err := http.PostForm(examples.APIURL, url.Values{
		`user`: {`xyzzy`},
	})
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

	if regexp.MustCompile(`^(?i)` + proxy.FullContentTypeHTML + `$`).
		MatchString(res.Header.Get(proxy.ContentTypeHeader)) {
		fmt.Println(string(body))
	} else {
		spew.Dump(body)
	}
	time.Sleep(700 * time.Millisecond)
}
