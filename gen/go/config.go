package gogen

type Config struct {
	root  string
	local string
	fn    string
}

func WithRoot(root string) func(*Config) {
	return func(c *Config) {
		c.root = lower(root)
	}
}

func WithLocal(local string) func(*Config) {
	return func(c *Config) {
		c.local = capitalize(lower(local))
	}
}

func WithFn(fn string) func(*Config) {
	return func(c *Config) {
		c.fn = lower(fn)
	}
}
