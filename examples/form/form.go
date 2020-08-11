package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/davecgh/go-spew/spew"

	"github.com/bearer/go-agent/examples"

	bearer "github.com/bearer/go-agent"
)

func main() {
	secretKey := os.Getenv(bearer.SecretKeyName)
	if len(secretKey) == 0 {
		log.Fatalf(`Bearer needs a %s environment variable`, bearer.SecretKeyName)
	}

	// Step 1: initialize Bearer.
	agent := bearer.New(secretKey)
	defer agent.Close()

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

	if regexp.MustCompile(`^(?i)text/html; charset=utf-8$`).MatchString(res.Header.Get(`Content-Type`)) {
		fmt.Println(string(body))
	} else {
		spew.Dump(body)
	}
}
