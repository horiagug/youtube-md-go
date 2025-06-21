package config

type Config struct {
	GeminiAPIKey string
	GeminiModel  string
	Timeout      int
	Languages    []string
}

func New(opts ...Option) *Config {
	cfg := &Config{
		Timeout: 60,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

type Option func(*Config)

func WithGeminiAPIKey(key string) Option {
	return func(c *Config) {
		c.GeminiAPIKey = key
	}
}

func WithLanguages(langs []string) Option {
	return func(c *Config) {
		c.Languages = langs
	}
}

func WithGeminiModel(key string) Option {
	return func(c *Config) {
		c.GeminiModel = key
	}
}
