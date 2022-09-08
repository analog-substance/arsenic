package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/d5/tengo/v2"
)

var scriptModule *ScriptModule = &ScriptModule{}

type ScriptModule struct {
	moduleMap  map[string]tengo.Object
	stopScript context.CancelFunc
}

func (m *ScriptModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"stop": &tengo.UserFunction{Name: "stop", Value: m.tengoStop},
		}
	}
	return m.moduleMap
}

func (m *ScriptModule) tengoStop(args ...tengo.Object) (tengo.Object, error) {
	message := ""
	if len(args) == 1 {
		message, _ = tengo.ToString(args[0])
	}

	m.stop(message)
	return nil, nil
}

func (m *ScriptModule) stop(args ...string) {
	if len(args) == 1 {
		fmt.Println(args[0])
	}

	go func() {
		m.stopScript()
	}()
	time.Sleep(1 * time.Millisecond)
}
