package agent_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	bearer "github.com/bearer/go-agent"
)

func TestE2E(t *testing.T) {
	secret := os.Getenv("BEARER_SECRETKEY")
	if secret == "" {
		t.Skip("TestE2E requires a valid $BEARER_SECRETKEY")
	}
	agent := bearer.New(secret)
	defer agent.Close()

	transport := agent.Decorate(&http.Transport{})
	client := http.Client{Transport: transport}

	resp, err := client.Get("https://httpbin.org/bytes/1000")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read body: %v", err)
	}

	if len(body) != 1000 {
		t.Errorf("invalid response length; expected %d, got %d.", 1000, len(body))
	}
}
