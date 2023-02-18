module github.com/analog-substance/arsenic

go 1.16

require (
	github.com/NoF0rte/gocdp v0.0.6
	github.com/Ullaakut/nmap/v2 v2.2.1
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/analog-substance/ffufwrap v0.0.0-20230214233527-0bbe7350af6d
	github.com/analog-substance/tengo/v2 v2.12.2
	github.com/andrew-d/go-termutil v0.0.0-20150726205930-009166a695a2
	github.com/bmatcuk/doublestar/v4 v4.2.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-playground/validator/v10 v10.10.1 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/reapertechlabs/go_nessus v0.1.2
	github.com/ryanuber/columnize v2.1.2+incompatible
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	github.com/ugorji/go v1.2.7 // indirect
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29 // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/Ullaakut/nmap/v2 v2.2.1 => github.com/analog-substance/nmap/v2 v2.2.2
