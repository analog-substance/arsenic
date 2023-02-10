package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/analog-substance/tengo/v2"
)

func (s *Script) ScriptModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"stop": &tengo.UserFunction{Name: "stop", Value: s.tengoStop},
	}
}

func (s *Script) tengoStop(args ...tengo.Object) (tengo.Object, error) {
	message := ""
	if len(args) == 1 {
		message, _ = tengo.ToString(args[0])
	}

	s.stop(message)
	return nil, nil
}

func (s *Script) stop(args ...string) {
	if len(args) == 1 {
		message := args[0]
		if message != "" {
			message = strings.ReplaceAll(message, `\n`, "\n")
			message = strings.ReplaceAll(message, `\t`, "\t")
			fmt.Println(message)
		}
	}

	go func() {
		s.cancel()
	}()
	time.Sleep(1 * time.Millisecond)
}
