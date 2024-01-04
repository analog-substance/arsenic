package log

import "fmt"

func Msg(args ...interface{}) {
	log("[+]", args...)
}
func Warn(args ...interface{}) {
	log("[!]", args...)
}
func Info(args ...interface{}) {
	log("[-]", args...)
}

func log(prefix string, args ...interface{}) {
	fmt.Printf("%s ", prefix)
	fmt.Print(args...)
	fmt.Println()
}
