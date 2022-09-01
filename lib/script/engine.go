package script

import (
	"context"

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
	m.stopScript()
	return nil, nil
}
