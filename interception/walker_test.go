package interception_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/bearer/go-agent/interception"
)

func TestWalker_WalkPreOrder(t *testing.T) {

	j1 := `
{
  "a":"1",
  "b":2,
  "sl":[
    "pre",
    "john.doe@example.com",
    "post",
    "fake370057577167325card"
  ],
  "ma":{
    "secret":[
      "pre-card",
      "fake370057577167325card",
      "post-card"
    ],
	"secret2": "fake370057577167325card",
    "foo":[
      "bar"
    ]
  }
}
`
	j2 := `5`

	tests := []struct {
		name string
		j    string
	}{
		{"typical", j1},
		{"degenerate: single int", j2},
	}

	p := interception.SanitizationProvider{
		SensitiveKeys:    []*regexp.Regexp{interception.DefaultSensitiveKeys},
		SensitiveRegexps: []*regexp.Regexp{interception.DefaultSensitiveData},
	}
	for _, tt := range tests {
		var x interface{}
		err := json.Unmarshal([]byte(tt.j), &x)
		if err != nil {
			t.Fatalf("unmarshalling test data: %v", err)
		}
		w := interception.NewWalker(x)

		var accu interface{}
		err = w.Walk(&accu, p.BodySanitizer)
		if err != nil {
			t.Errorf("WalkPreOrder error: %v", err)
		}
		fmt.Println(w)
	}
}
