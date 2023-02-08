package engine

import (
	"github.com/analog-substance/tengo/v2"
)

// Warning represents a warning value.
type Warning struct {
	tengo.ObjectImpl
	Value tengo.Object
}

// TypeName returns the name of the type.
func (o *Warning) TypeName() string {
	return "warning"
}

func (o *Warning) String() string {
	if o.Value != nil {
		return o.Value.String()
	}
	return "warning"
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Warning) IsFalsy() bool {
	return true // warning is always false.
}

// Copy returns a copy of the type.
func (o *Warning) Copy() tengo.Object {
	return &Warning{Value: o.Value.Copy()}
}

// Equals returns true if the value of the type is equal to the value of
// another object.
func (o *Warning) Equals(x tengo.Object) bool {
	return o == x // pointer equality
}

// IndexGet returns an element at a given index.
func (o *Warning) IndexGet(index tengo.Object) (res tengo.Object, err error) {
	if strIdx, _ := tengo.ToString(index); strIdx != "value" {
		err = tengo.ErrInvalidIndexOnError
		return
	}
	res = o.Value
	return
}
