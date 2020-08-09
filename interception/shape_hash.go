//go:generate protoc -I=$PWD/interception --go_out=$PWD/interception $PWD/interception/shape_hash.proto

package interception

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"

	mini "github.com/tdewolff/minify/v2"
	miniJ "github.com/tdewolff/minify/v2/json"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/bearer/go-agent/proxy"
)

var minifier *mini.M

// NewShapeDescriptor builds a new ShapeDescriptor from its fields.
func NewShapeDescriptor(typ ShapeDescriptor_PrimitiveType, fields []*FieldDescriptor, items []*ShapeDescriptor) *ShapeDescriptor {
	if fields == nil {
		fields = make([]*FieldDescriptor, 0)
	}
	if items == nil {
		items = make([]*ShapeDescriptor, 0)
	}
	return &ShapeDescriptor{
		Type:   typ,
		Fields: fields,
		Items:  items,
	}
}

func jsonToShapeHash(x interface{}) (*ShapeDescriptor, error) {
	var ret *ShapeDescriptor
	typ := reflect.TypeOf(x)
	var kind reflect.Kind
	if typ == nil {
		kind = reflect.Invalid
	} else {
		kind = typ.Kind()
	}
	switch kind {
	case reflect.Slice:
		sl := reflect.ValueOf(x)
		items := make([]*ShapeDescriptor, sl.Len())
		for i := 0; i < sl.Len(); i++ {
			h, err := jsonToShapeHash(sl.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			items[i] = h
		}
		ret = NewShapeDescriptor(ShapeDescriptor_ARRAY, nil, items)

	case reflect.Map:
		// Go maps iterate in pseudo-random order, regardless of insertion order,
		// so we sort the keys separately.
		ma := reflect.ValueOf(x)
		var keys sort.StringSlice = make([]string, ma.Len())
		for i, rKey := range ma.MapKeys() {
			if rKey.Kind() != reflect.String {
				return nil, fmt.Errorf(`non-string key %v in map`, rKey.Interface())
			}
			keys[i] = rKey.String()
		}
		keys.Sort()

		fields := make([]*FieldDescriptor, len(keys))
		for i, key := range keys {
			fields[i] = &FieldDescriptor{Key: key}
			v := ma.MapIndex(reflect.ValueOf(key)).Interface()
			h, err := jsonToShapeHash(v)
			if err != nil {
				return nil, fmt.Errorf(`could not shape field %s: %v`, key, err)
			}
			fields[i].Hash = h
		}
		ret = NewShapeDescriptor(ShapeDescriptor_OBJECT, fields, nil)

	case reflect.Int, reflect.Uintptr,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		ret = NewShapeDescriptor(ShapeDescriptor_NUMBER, nil, nil)

	case reflect.String:
		ret = NewShapeDescriptor(ShapeDescriptor_STRING, nil, nil)

	case reflect.Invalid:
		ret = NewShapeDescriptor(ShapeDescriptor_NULL, nil, nil)

	case reflect.Bool:
		ret = NewShapeDescriptor(ShapeDescriptor_BOOLEAN, nil, nil)
	default:
		return nil, fmt.Errorf(`unknown type! %T`, x)
	}

	return ret, nil
}

// ToBytes builds a hex-encoded representation of the shape of its argument.
func ToBytes(x interface{}) ([]byte, error) {
	hashMessage, err := jsonToShapeHash(x)
	if err != nil {
		return nil, err
	}
	mo := protojson.MarshalOptions{
		Multiline:       false,
		Indent:          ``,
		UseEnumNumbers:  true,
		EmitUnpopulated: true,
	}
	j, err := mo.Marshal(hashMessage)
	if err != nil {
		return nil, err
	}
	// This output is pseudo-random: Go goes out of its way to ensure the
	// marshalling result changes from one build to the next, so we have to
	// minify it after the fact to ensure a decent probability of format
	// consistency.
	//
	// See https://github.com/golang/protobuf/issues/1121 and package
	// internal/detrand in protobuf for the implementation.
	j, err = minifier.Bytes(proxy.ContentTypeJSON, j)
	return j, err
}

// ToHash builds a NewShapeDescriptor of its argument.
func ToHash(j interface{}) string {
	bytes, err := ToBytes(j)
	if err != nil {
		bytes = nil
	}
	return hex.EncodeToString(bytes)
}

// ToSha builds a SHA256 of the NewShapeDescriptor of its argument.
func ToSha(j interface{}) string {
	bytes, err := ToBytes(j)
	if err != nil {
		return `N/A`
	}
	sha := sha256.Sum256(bytes)
	return hex.EncodeToString(sha[:])
}

func init() {
	minifier = mini.New()
	minifier.AddFunc(proxy.ContentTypeJSON, miniJ.Minify)
}
