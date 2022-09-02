package script

import (
	"context"
	"fmt"
	"time"

	"github.com/d5/tengo/v2"
)

var engineModule *EngineModule = &EngineModule{}

type EngineModule struct {
	moduleMap  map[string]tengo.Object
	stopScript context.CancelFunc
}

func (m *EngineModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"stop": &tengo.UserFunction{Name: "stop", Value: m.stop},
		}
	}
	return m.moduleMap
}

func (m *EngineModule) stop(args ...tengo.Object) (tengo.Object, error) {
	message := ""
	if len(args) == 1 {
		message, _ = tengo.ToString(args[0])
	}

	stopScript(message)
	return nil, nil
}

func stopScript(args ...string) {
	if len(args) == 1 {
		fmt.Println(args[0])
	}

	go func() {
		engineModule.stopScript()
	}()
	time.Sleep(1 * time.Millisecond)
}
