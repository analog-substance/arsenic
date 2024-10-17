package pkg

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/analog-substance/arsenic/pkg/config"
	"github.com/analog-substance/util/set"
	"github.com/spf13/viper"
)

func Test_cleanLine(t *testing.T) {
	type args struct {
		wordlistType string
		line         string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Should not clean pizza",
			args{"web-content", "pizza"},
			"pizza",
		},
		{
			"Should clean forward slash from /pizza",
			args{"web-content", "/pizza"},
			"pizza",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanLine(tt.args.wordlistType, tt.args.line); got != tt.want {
				t.Errorf("cleanLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldIgnoreLine(t *testing.T) {
	wordlists := config.Wordlists{
		Types: map[string][]string{
			"web-content": {},
			"sqli":        {},
			"xss":         {},
		},
	}
	setConfigDefault("wordlists", wordlists)

	var c config.Config
	viper.Unmarshal(&c)
	config.Set(&c)

	type args struct {
		wordlistType string
		line         string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Should not ignore /pizza",
			args{"web-content", "/pizza"},
			false,
		},
		{
			"Should ignore comments #pizza",
			args{"web-content", "#pizza"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldIgnoreLine(tt.args.wordlistType, tt.args.line); got != tt.want {
				t.Errorf("shouldIgnoreLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readWordlist(t *testing.T) {
	type args struct {
		wordlistType string
		lineSet      set.Set
		reader       bytes.Buffer
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"Should read wordlist",
			args{"web-content", set.NewSet(""), *bytes.NewBufferString("one\ntwo\n")},
			2,
		},
		{
			"Should read wordlist and remove dupes",
			args{"web-content", set.NewSet(""), *bytes.NewBufferString("one\ntwo\none\n")},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readWordlist(tt.args.wordlistType, &tt.args.lineSet, &tt.args.reader)
			if got := tt.args.lineSet.Length(); got != tt.want {
				t.Errorf("readWordlist(), Length() %v, want %v", got, tt.want)
			}
		})
	}
}

// If no config file exists, all possible keys in the defaults
// need to be registered with viper otherwise viper will only think
// the keys explicitly set via viper.SetDefault() exist.
func setConfigDefault(key string, value interface{}) {
	valueType := reflect.TypeOf(value)
	valueValue := reflect.ValueOf(value)

	if valueType.Kind() == reflect.Map {
		iter := valueValue.MapRange()
		for iter.Next() {
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			setConfigDefault(fmt.Sprintf("%s.%s", key, k), v)
		}
	} else if valueType.Kind() == reflect.Struct {
		numFields := valueType.NumField()
		for i := 0; i < numFields; i++ {
			structField := valueType.Field(i)
			fieldValue := valueValue.Field(i)

			if !structField.IsExported() {
				continue
			}

			subKey := structField.Name
			yamlTag := structField.Tag.Get("yaml")
			if yamlTag == "-" {
				continue
			}

			if yamlTag != "" {
				subKey = yamlTag
			}

			fullKey := fmt.Sprintf("%s.%s", key, subKey)
			if key == "" {
				fullKey = subKey
			}

			setConfigDefault(fullKey, fieldValue.Interface())
		}
	} else {
		viper.SetDefault(key, value)
	}
}
