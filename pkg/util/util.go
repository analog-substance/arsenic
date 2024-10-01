package util

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
)

func StringSliceEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func GetReviewer(reviewerFlag string) string {
	if reviewerFlag == "operator" {
		envReviewer := os.Getenv("AS_REVIEWER")
		envUser := os.Getenv("USER")
		if len(envReviewer) > 0 {
			reviewerFlag = envReviewer
		} else if len(envUser) > 0 {
			reviewerFlag = envUser
		}
	}

	return reviewerFlag
}

type NoopWriter struct {
}

func (w NoopWriter) Write(bytes []byte) (int, error) {
	return 0, nil
}

func ToString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
func ToStringSlice(v interface{}) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, ToString(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, ToString(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{ToString(v)}
		}
	}
}

func IndexOf(data []string, item string) int {
	for k, v := range data {
		if item == v {
			return k
		}
	}
	return -1
}

func RemoveIndex(arr []string, idx int) []string {
	return append(arr[:idx], arr[idx+1:]...)
}

func Sanitize(s string) string {
	// Windows is most restrictive
	windows_regex := regexp.MustCompile("[<>:/\\|?*\"]+")
	s = windows_regex.ReplaceAllString(s, "_")
	return strings.TrimSpace(s)
}

func IsIp(ipOrHostname string) bool {
	if net.ParseIP(ipOrHostname) == nil {
		return false
	} else {
		return true
	}
}
