package set

import (
	"reflect"
	"sort"

	"github.com/ahmetb/go-linq/v3"
)

type Set struct {
	Set      map[interface{}]bool
	itemType reflect.Type
}

func NewSet(itemType reflect.Type) *Set {
	return &Set{itemType: itemType, Set: map[interface{}]bool{}}
}
func NewStringSet() *Set {
	return NewSet(reflect.TypeOf(""))
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
	linq.From(items).ForEach(func(i interface{}) { set.Set[i] = true })
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
func (set *Set) StringSlice() []string {
	if set.itemType != reflect.TypeOf("") {
		return nil
	}
	return set.Slice().([]string)
}
func (set *Set) SortedStringSlice() []string {
	if set.itemType != reflect.TypeOf("") {
		return nil
	}
	slice := set.StringSlice()
	sort.Strings(slice)
	return slice
}
func (set *Set) Length() int {
	return len(set.Set)
}
