package script

import "github.com/d5/tengo/v2"

var gitModule = map[string]tengo.Object{
	"pull":   &tengo.UserFunction{Name: "pull", Value: gitPull},
	"commit": &tengo.UserFunction{Name: "commit", Value: gitCommit},
	"lock":   &tengo.UserFunction{Name: "lock", Value: gitLock},
}

func gitPull(args ...tengo.Object) (tengo.Object, error) {

	return nil, nil
}
func gitCommit(args ...tengo.Object) (tengo.Object, error) {
	return nil, nil
}
func gitLock(args ...tengo.Object) (tengo.Object, error) {
	return nil, nil
}
