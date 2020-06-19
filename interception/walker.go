package interception

import (
	"fmt"
	"reflect"
)

// WalkFn is the type for visitor functions used with a Walker.
type WalkFn func(ik interface{}, iv *interface{}, accu *interface{}) error

// Walker is able to walk a visitor WalkFn in preorder across the whole tree of
// a value unmarshalled from JSON, which is far from being any type of Go data.
type Walker interface {
	fmt.Stringer
	Walk(accu *interface{}, visitor WalkFn) error
	Value() interface{}
}

// NewWalker builds an initialized Walker.
func NewWalker(x interface{}) Walker {
	return walker{
		root: x,
	}
}

type walker struct {
	root interface{}
}

func (w walker) String() string {
	return fmt.Sprint(w.root)
}

func (w walker) Value() interface{} {
	return w.root
}

func (w walker) Walk(accu *interface{}, visitor WalkFn) error {
	return w.walkPreOrder(nil, &w.root, accu, visitor)
}

func (w walker) walkPreOrder(k interface{}, v *interface{}, accu *interface{}, visitor WalkFn) error {
	if err := visitor(k, v, accu); err != nil {
		return err
	}

	value := reflect.ValueOf(*v)
	typ := reflect.TypeOf(*v)
	var kind reflect.Kind
	if typ == nil {
		kind = reflect.Invalid
	} else {
		kind = typ.Kind()
	}
	switch kind {
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			vi := v.Interface()
			err := w.walkPreOrder(k.Interface(), &vi, accu, visitor)
			if err != nil {
				return err
			}
			value.SetMapIndex(k, reflect.ValueOf(vi))
		}
	case reflect.Slice:
		len := value.Len()
		for i := 0; i < len; i++ {
			v := value.Index(i)
			vi := v.Interface()
			if err := w.walkPreOrder(i, &vi, accu, visitor); err != nil {
				return err
			}
			v.Set(reflect.ValueOf(vi))
		}
	default:
		// Nothing to walk.
	}

	return nil
}

