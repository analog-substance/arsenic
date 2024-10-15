package set

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"io"
	"os"
	"reflect"
	"sort"
)

type Set struct {
	Set      map[interface{}]bool
	itemType reflect.Type
}

func NewSet(itemType interface{}) Set {
	return Set{
		itemType: reflect.TypeOf(itemType),
		Set:      map[interface{}]bool{},
	}
}

func NewStringSet(values ...[]string) *Set {
	s := NewSet("")
	for _, value := range values {
		s.AddRange(value)
	}
	return &s
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
	linq.From(items).ForEach(func(i interface{}) { set.Add(i) })
}

func (set *Set) Contains(item interface{}) bool {
	_, found := set.Set[item]
	return found
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

func (set *Set) WriteSorted(writer io.Writer) {
	for _, line := range set.SortedStringSlice() {
		fmt.Fprintln(writer, line)
	}
}

func (set *Set) PrintSorted() {
	set.WriteSorted(os.Stdout)
}
