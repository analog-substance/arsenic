package slice

import "reflect"

// Any determines whether any element of a sequence satisfies a condition.
func Any(x interface{}, predicate func(item interface{}) bool) bool {
	xValue := reflect.ValueOf(x)
	if xValue.Kind() != reflect.Slice {
		return false
	}

	length := xValue.Len()
	for i := 0; i < length; i++ {
		value := xValue.Index(i).Interface()
		if predicate(value) {
			return true
		}
	}
	return false
}

// ForEach performs the specified action on each element of the specified slice.
func ForEach(x interface{}, action func(item interface{})) {
	xValue := reflect.ValueOf(x)
	if xValue.Kind() != reflect.Slice {
		return
	}

	length := xValue.Len()
	for i := 0; i < length; i++ {
		value := xValue.Index(i).Interface()
		action(value)
	}
}

// Filter filters a sequence of values based on a predicate.
func Filter(x interface{}, predicate func(item interface{}) bool) interface{} {
	xValue := reflect.ValueOf(x)
	if xValue.Kind() != reflect.Slice {
		return false
	}

	length := xValue.Len()
	values := reflect.MakeSlice(xValue.Type(), 0, 0)
	for i := 0; i < length; i++ {
		value := xValue.Index(i).Interface()
		if predicate(value) {
			values = reflect.Append(values, reflect.ValueOf(value))
		}
	}
	return values.Interface()
}

// Map projects each element of a sequence into a new form.
func Map(x interface{}, resultType reflect.Type, mapper func(item interface{}) interface{}) interface{} {
	xValue := reflect.ValueOf(x)
	if xValue.Kind() != reflect.Slice {
		return nil
	}

	length := xValue.Len()
	values := reflect.MakeSlice(resultType, 0, 0)
	for i := 0; i < length; i++ {
		value := xValue.Index(i).Interface()
		newValue := reflect.ValueOf(mapper(value))
		if newValue.Type().Name() != resultType.Name() {
			continue
		}
		values = reflect.Append(values, newValue)
	}
	return values.Interface()
}
