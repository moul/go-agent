package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bearer/go-agent"
	"github.com/bearer/go-agent/examples"
	"github.com/bearer/go-agent/proxy"
)

func main() {
	secretKey := os.Getenv(agent.SecretKeyName)
	if len(secretKey) == 0 {
		log.Fatalf(`Bearer needs a %s environment variable`, agent.SecretKeyName)
	}

	// Step 1: initialize Bearer.
	defer agent.Init(secretKey)()

	// Step 2: prepare your custom transport, decorating it with the Bearer agent.
	var baseTransport = http.DefaultTransport.(*http.Transport)

	// Say your enterprise proxy needs a specific CONNECT header.
	baseTransport.ProxyConnectHeader = http.Header{"ACME_ID": []string{"some secret"}}
	transport := agent.DefaultAgent.Decorate(baseTransport)

	// Step 3: use your client as usual.
	client := http.Client{Transport: transport}
	res, err := client.Do(&http.Request{
		Method: http.MethodGet,
		URL:    proxy.MustParseURL(examples.APIURL),
	})
	if err != nil {
		log.Fatalf("calling %s: %v", examples.APIURL, err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("reading API response: %v", err)
	}

	examples.ShowGithubOrg(body)
	time.Sleep(6 * time.Second)
}
