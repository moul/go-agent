package interception

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

var (
	spongeBob = map[string]interface{}{
		`name`:    `Sponge Bob`,
		`age`:     12,
		`friends`: []interface{}{`patrick`, `mr krab`, `starman`},
	}
)

func TestToHash(t *testing.T) {
	tests := []struct {
		name  string
		x     map[string]interface{}
		equal bool
	}{
		{"NPM 1", map[string]interface{}{`firstname`: `Aidan`, `homestate`: `PA`}, false},
		{"NPM 2", map[string]interface{}{`firstname`: `Dev`, `homestate`: `New Jersey`}, false},
		{"Ruby Patrick", map[string]interface{}{
			`name`:    `Patrick`,
			`age`:     5,
			`friends`: []interface{}{`Sponge Bob`, `mr krab`, `starman`},
		}, true},
		{`Squarepants`, spongeBob, true},
		{`Ruby Starman`, map[string]interface{}{
			`name`:    `Starman`,
			`age`:     5,
			`friends`: []interface{}{`Sponge Bob`, `Patrick`},
		}, false},
		{`Ruby Superman`, map[string]interface{}{
			`name`:     `Superman`,
			`age`:      79,
			`powers`:   []interface{}{`flying`},
			`location`: nil,
			`animal`:   false,
		}, false},
	}
	hSB := ToHash(spongeBob)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHash(tt.x)

			actualEqual := got == hSB
			if actualEqual != tt.equal {
				t.Errorf(`toHash(%s) equal ? want %t, got %t`, tt.name, tt.equal, actualEqual)
			}
		})
	}
}

func TestToHashValue(t *testing.T) {
	// Cf. Ruby Agent.
	const expected = `7b226669656c6473223a5b7b2268617368223a7b226669656c6473223a5b5d2c226974656d73223a5b5d2c2272756c6573223a5b5d2c2274797065223a337d2c226b6579223a22616765227d2c7b2268617368223a7b226669656c6473223a5b5d2c226974656d73223a5b7b226669656c6473223a5b5d2c226974656d73223a5b5d2c2272756c6573223a5b5d2c2274797065223a327d2c7b226669656c6473223a5b5d2c226974656d73223a5b5d2c2272756c6573223a5b5d2c2274797065223a327d2c7b226669656c6473223a5b5d2c226974656d73223a5b5d2c2272756c6573223a5b5d2c2274797065223a327d5d2c2272756c6573223a5b5d2c2274797065223a317d2c226b6579223a22667269656e6473227d2c7b2268617368223a7b226669656c6473223a5b5d2c226974656d73223a5b5d2c2272756c6573223a5b5d2c2274797065223a327d2c226b6579223a226e616d65227d5d2c226974656d73223a5b5d2c2272756c6573223a5b5d2c2274797065223a307d`
	actual := ToHash(spongeBob)

	if testing.Verbose() {
		sExp, _ := hex.DecodeString(expected)
		fmt.Printf("%s\n", sExp)
		sAct, _ := hex.DecodeString(actual)
		fmt.Printf("%s\n", sAct)
	}

	if len(actual) != len(expected) {
		t.Errorf(`ToSha(spongebob) got %d bytes, expected %d`, len(actual), len(expected))
	}
	if actual != expected {
		t.Errorf(`ToHash(spongebob) got %s, expected %s`, actual, expected)
	}
}

func TestToSha(t *testing.T) {
	// Cf. Ruby Agent.
	const expected = `9d50c0ee5be33590542a35b92f4bfef7770aae21927d4ba8f4804fb108cb3b55`
	actual := ToSha(spongeBob)
	if actual != expected {
		t.Errorf(`ToSha(spongebob) got %s, expected %s`, actual, expected)
	}
}

func TestToBytes(t *testing.T) {
	tests := []struct {
		name    string
		x       interface{}
		want    []byte
		wantErr bool
	}{
		{"happy string", "foo",
			[]byte(`{"fields":[],"items":[],"rules":[],"type":2}`),
			false},
		{"happy slice", []interface{}{"foo", 5},
			[]byte(`{"fields":[],"items":[{"fields":[],"items":[],"rules":[],"type":2},{"fields":[],"items":[],"rules":[],"type":3}],"rules":[],"type":1}`),
			false},
		{"happy nil", nil,
			[]byte(`{"fields":[],"items":[],"rules":[],"type":5}`), false},
		{"sad bool map", map[bool]bool{true: false}, nil, true},
		{"happy map", map[string]bool{"true": false}, []byte(`{"fields":[{"hash":{"fields":[],"items":[],"rules":[],"type":4},"key":"true"}],"items":[],"rules":[],"type":0}`),
			false},
		{"sad func", []interface{}{func() {}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToBytes(tt.x)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToBytes() got = %s, want %s", got, tt.want)
			}
		})
	}
}
