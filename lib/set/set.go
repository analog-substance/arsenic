package set

import (
	"reflect"

	"github.com/defektive/arsenic/lib/slice"
)

type Set struct {
	Set      map[interface{}]bool
	itemType reflect.Type
}

func NewSet(itemType reflect.Type) Set {
	return Set{itemType: itemType, Set: map[interface{}]bool{}}
}
func (set *Set) Add(item interface{}) bool {
	itemType := reflect.TypeOf(item)
	if itemType.Name() != set.itemType.Name() {
		return false
	}

	_, found := set.Set[item]
	set.Set[item] = true
	return !found
}
func (set *Set) AddRange(items interface{}) {
	slice.ForEach(items, func(item interface{}) { set.Set[item] = true })
}
func (set *Set) Slice() interface{} {
	sliceType := reflect.SliceOf(set.itemType)
	values := reflect.MakeSlice(sliceType, 0, len(set.Set))

	for s := range set.Set {
		values = reflect.Append(values, reflect.ValueOf(s))
	}
	rawValues := values.Interface()
	return rawValues
}
