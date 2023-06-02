package nmaputil

type Options struct {
	upOnly   bool
	openOnly bool
}

type Option func(*Options)

func WithUpOnly() Option {
	return func(o *Options) {
		o.upOnly = true
	}
}

func WithOpenOnly() Option {
	return func(o *Options) {
		o.openOnly = true
	}
}
