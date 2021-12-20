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
			args{"one"},
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
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}

			set.AddRange([]string{"one", "two"})
		})
	}
}

func TestSet_Length(t *testing.T) {
	type fields struct {
		Set      map[interface{}]bool
		itemType reflect.Type
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Set:      tt.fields.Set,
				itemType: tt.fields.itemType,
			}
			if got := set.Length(); got != tt.want {
				t.Errorf("Length() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
