package set

import (
	"bytes"
	"reflect"
	"testing"
)

func TestSet_Add(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Test add string to string set",
			fields{map[interface{}]bool{}, reflect.TypeOf("")},
			args{"one"},
			true,
		},
		{
			"Test add int to string set",
			fields{map[interface{}]bool{}, reflect.TypeOf("")},
			args{1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}
			if got := set.Add(tt.args.item); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_AddRange(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	type args struct {
		items interface{}
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		desiredLengthBefore int
		desiredLengthAfter  int
	}{
		{
			"Test add string to string set",
			fields{map[interface{}]bool{}, reflect.TypeOf("")},
			args{[]string{"one", "two"}},
			0,
			2,
		},
		{
			"Test add int to string set",
			fields{map[interface{}]bool{"one": true}, reflect.TypeOf("")},
			args{[]string{"one", "two"}},
			1,
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}

			if got := set.Length(); got != tt.desiredLengthBefore {
				t.Errorf("Before AddRange() then Length() = %v, want %v", got, tt.desiredLengthBefore)
			}

			set.AddRange(tt.args.items)

			if got := set.Length(); got != tt.desiredLengthAfter {
				t.Errorf("AddRange() then Length() = %v, want %v", got, tt.desiredLengthAfter)
			}

		})
	}
}

func TestSet_Slice(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			"Test get int slice",
			fields{map[interface{}]bool{1: true}, reflect.TypeOf(1)},
			[]int{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}
			if got := set.Slice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_SortedStringSlice(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			"Test get string slice",
			fields{map[interface{}]bool{"ddd": true, "aaa": true}, reflect.TypeOf("")},
			[]string{"aaa", "ddd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}
			if got := set.SortedStringSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortedStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_StringSlice(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			"Test get string slice",
			fields{map[interface{}]bool{"one": true}, reflect.TypeOf("")},
			[]string{"one"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}
			if got := set.StringSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_WriteSorted(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	tests := []struct {
		name       string
		fields     fields
		wantWriter string
	}{
		{
			"Test writing sorted",
			fields{map[interface{}]bool{"ddd": true, "aaa": true}, reflect.TypeOf("")},
			"aaa\nddd\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}
			writer := &bytes.Buffer{}
			set.WriteSorted(writer)
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("WriteSorted() = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
