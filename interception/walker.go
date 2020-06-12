package interception

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type WalkFn func(ik interface{}, iv *interface{}) error

type Walker interface {
	fmt.Stringer
	Walk(WalkFn) error
}

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

func (w walker) Walk(fn WalkFn) error {
	return w.walkPreOrder(nil, &w.root, fn)
}

func (w walker) walkPreOrder(k interface{}, v *interface{}, fn WalkFn) error {
	if err := fn(k, v); err != nil {
		return err
	}

	value := reflect.ValueOf(*v)
	kind := reflect.TypeOf(*v).Kind()
	switch kind {
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			vi := v.Interface()
			err := w.walkPreOrder(k.Interface(), &vi, fn)
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
			if err := w.walkPreOrder(i, &vi, fn); err != nil {
				return err
			}
			v.Set(reflect.ValueOf(vi))
		}
	default:
		// Nothing to walk.
	}

	return nil
}

func BodySanitizer(k interface{}, v *interface{}) error {
	s := fmt.Sprintf("%v %v", k, *v)
	log.Print(s)
	if k == nil {
		return nil
	}
	if sk, ok := k.(string); ok {
		if strings.Contains(sk, "secret") {
			*v = Filtered
			return nil
		}
	}

	if reflect.ValueOf(*v).Kind() == reflect.String {
		sv, ok := (*v).(string)
		if !ok {
			return fmt.Errorf("not a string: %#v", *v)
		}
		if strings.Contains(sv, "card") {
			sv = strings.Replace(sv, "card", Filtered, -1)
		}
		*v = sv
	}
	return nil
}
