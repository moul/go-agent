package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

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

	// Step 2: prepare your custom transport.
	// eg. your enterprise proxy needs a specific CONNECT header
	baseTransport := &http.Transport{}
	baseTransport.ProxyConnectHeader = http.Header{"ACME_ID": []string{"some secret"}}

	// Step 3: decorate your transport with Bearer.
	transport := agent.Decorate(baseTransport)

	// Step 4: use your client as usual.
	client := http.Client{Transport: transport}
	res, err := client.Get(examples.APIURL)
	if err != nil {
		log.Fatalf("calling %s: %v", examples.APIURL, err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("reading API response: %v", err)
	}

	examples.ShowGithubOrg(body)
}
