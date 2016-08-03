// Package hy is a two-way hierarchical YAML parser.
//
// hy allows you to read and write YAML files in a directory hierarchy
// to and from go structs. It uses tags to define the locations of
// YAML files and directories containing YAML files, and some simple
// mapping of filenames to string values, used during packing and unpacking.
package hy

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/opentable/sous/util/yaml"
)

// Error wraps YAML errors in order to indicate which file contains them
type Error struct {
	file  string
	cause error
}

// Unmarshaler tracks the process of deserializing a directory of yaml files
// into a set of structs
type Unmarshaler struct {
	UnmarshalFunc func([]byte, interface{}) error
}

// NewUnmarshaler creates an Unmarshaler
func NewUnmarshaler(unmarshalFunc func([]byte, interface{}) error) Unmarshaler {
	if unmarshalFunc == nil {
		panic("unmarshalFunc must not be nil")
	}
	return Unmarshaler{unmarshalFunc}
}

// Unmarshal is shorthand for NewUnmarshaler(yaml.Unmarshal).Unmarshal
func Unmarshal(path string, v interface{}) error {
	return NewUnmarshaler(yaml.Unmarshal).Unmarshal(path, v)
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.file, e.cause)
}

// Unmarshal deserializes from a directory
func (u Unmarshaler) Unmarshal(path string, v interface{}) error {
	if v == nil {
		return fmt.Errorf("hy cannot unmarshal to nil")
	}
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !s.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return ctx{path, u.UnmarshalFunc, nil}.unmarshalDir(v)
}

func (c ctx) unmarshalDir(v interface{}) error {
	targets, err := c.getStructTargets(v)
	if err != nil {
		return err
	}
	return targets.unmarshalAll(nil)
}

func (ts targets) unmarshalAll(parent *reflect.Value) error {
	for _, t := range ts {
		if err := t.unmarshal(parent); err != nil {
			if _, ok := err.(*Error); ok {
				return err
			}
			if os.IsNotExist(err) || os.IsPermission(err) {
				return err
			}
			return &Error{t.path, err}
		}
		debug(t.val.Type(), t.val.Interface())
		if parent == nil {
			debug(parent, "\n")
		} else if parent.Kind() == reflect.Struct {
			debug(parent, parent.FieldByName(t.name), "\n")
		} else if parent.Kind() == reflect.Ptr && parent.Elem().Kind() == reflect.Struct {
			debug(parent, parent.Elem().FieldByName(t.name), "\n")
		} else {
			debug(parent, "\n")
		}

	}
	return nil
}

func (t target) unmarshal(parent *reflect.Value) error {
	debugf("Target: %s\n", t.path)
	iface := t.val.Interface()
	if isFile(t.path) {
		debug("unmarshall file", t)
		if err := t.unmarshalFile(iface); err != nil {
			return err
		}
	}
	if len(t.subTargets) != 0 {
		debug("subtargets", t)
		if err := t.subTargets.unmarshalAll(&t.val); err != nil {
			return err
		}
	}
	if parent != nil {
		debug("insert into parent", parent, t)
		if err := t.insertIntoParent(parent); err != nil {
			return err
		}
	}
	debug(t.val.Interface(), "\n")
	return nil
}

func parentTypeError(parent *reflect.Value) error {
	return fmt.Errorf("parent was %s; want pointer or map[string]T", parent.Type())
}

func (t target) insertIntoParent(parent *reflect.Value) error {
	debug(parent, parent.Kind())
	switch parent.Kind() {
	default:
		return parentTypeError(parent)
	case reflect.Ptr:
		if parent.Elem().Kind() != reflect.Struct {
			return parentTypeError(parent)
		}
		debugf("Setting field %s on %s\n", t.name, parent.Elem().Type())
		f := parent.Elem().FieldByName(t.name)
		f.Set(*getConcreteValRef(t.val))
	case reflect.Map:
		if parent.Type().Key().Kind() != reflect.String {
			return parentTypeError(parent)
		}
		debugf("Setting key %q on %s\n", t.name, parent.Type())
		if parent.IsNil() {
			pvp := reflect.MakeMap(parent.Type())
			debugf("Parent was nil, setting empty map of %s\n", parent.Type())
			parent.Set(pvp)
		}
		elem := t.val
		if parent.Type().Elem().Kind() != reflect.Ptr {
			elem = elem.Elem()
		}
		parent.SetMapIndex(reflect.ValueOf(t.name), elem)
		debug(parent.Interface())
	}
	return nil
}

func (t target) unmarshalFile(iface interface{}) error {
	if t.val.Kind() != reflect.Ptr || t.val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("tried to unmarshal file %s to %T; want a pointer to struct", t.path, iface)
	}
	debugf("Path: %s; Type: %s; ValType: %s; IfaceType: %s\n", t.path, t.typ, t.val.Type(), reflect.TypeOf(t.val.Interface()))
	b, err := ioutil.ReadFile(t.path)
	if err != nil {
		return err
	}
	if err := t.unmarshalFunc(b, iface); err != nil {
		return err
	}
	debugf("Unmarshalled: val: %v; type: %T", iface, iface)
	return nil
}

// getElemType tries to get element type of a map or slice, and if that type is
// a pointer, gets the element type of the pointer instead. If the elem type is
// pointer to pointer, or the type passed in is not a map or slice,  returns an
// error.
func getElemType(typ reflect.Type) (reflect.Type, error) {
	k := typ.Kind()
	switch k {
	default:
		return nil, fmt.Errorf("directory target not allowed for type %s; want map or slice", typ)
	case reflect.Slice, reflect.Map:
		break
	}
	elemType := typ.Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Ptr {
		return nil, fmt.Errorf("%s containing %s not supported", k, typ.Elem())
	}
	return elemType, nil
}

func newValue(typ reflect.Type) reflect.Value {
	if typ.Kind() == reflect.Ptr {
		panic("newValue passed a pointer type")
	}
	return reflect.New(typ).Elem()
}

func getConcreteValRef(v reflect.Value) *reflect.Value {
	v = getConcreteVal(&v)
	return &v
}

func getConcreteVal(v *reflect.Value) reflect.Value {
	switch v.Kind() {
	default:
		return *v
	case reflect.Ptr:
		e := v.Elem()
		return e
	}
}

func isFile(path string) bool {
	return strings.HasSuffix(path, ".yaml")
}
