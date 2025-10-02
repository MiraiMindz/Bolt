package bolt

import "time"

// DefaultConfig returns the default configuration  
func DefaultConfig() Config {  
    return Config{  
        ReadTimeout:       15 * time.Second,  
        WriteTimeout:      15 * time.Second,  
        IdleTimeout:       60 * time.Second,  
        GenerateDocs:      true,  
        EnablePooling:     true,  
        MaxPoolSize:       1000,  
        PreallocateRoutes: 100,  
        DevMode:           false,  
        DocsConfig: DocsConfig{  
            Enabled:     true,  
            SpecPath:    "/openapi.json",  
            UIPath:      "/docs",  
            Title:       "API Documentation",  
            Description: "Automatically generated API documentation",  
            Version:     "1.0.0",  
        },  
    }  
}

// WithDocs enables or disables documentation  
func WithDocs(enabled bool) Option {  
    return func(c *Config) {  
        c.GenerateDocs = enabled  
        c.DocsConfig.Enabled = enabled  
    }  
}

// WithDocsPath sets documentation paths  
func WithDocsPath(specPath, uiPath string) Option {  
    return func(c *Config) {  
        c.DocsConfig.SpecPath = specPath  
        c.DocsConfig.UIPath = uiPath  
    }  
}

// WithAPIInfo sets API metadata  
func WithAPIInfo(title, description, version string) Option {  
    return func(c *Config) {  
        c.DocsConfig.Title = title  
        c.DocsConfig.Description = description  
        c.DocsConfig.Version = version  
    }  
}

// WithDevMode enables development mode  
func WithDevMode(enabled bool) Option {  
    return func(c *Config) {  
        c.DevMode = enabled  
    }  
}

// WithTimeouts sets server timeouts  
func WithTimeouts(read, write, idle time.Duration) Option {  
    return func(c *Config) {  
        c.ReadTimeout = read  
        c.WriteTimeout = write  
        c.IdleTimeout = idle  
    }  
}

// WithPooling enables or disables object pooling  
func WithPooling(enabled bool) Option {  
    return func(c *Config) {  
        c.EnablePooling = enabled  
    }  
}

// WithCustomDocGenerator sets a custom documentation generator  
func WithCustomDocGenerator(gen DocGenerator) Option {  
    return func(c *Config) {  
        c.DocsConfig.Generator = gen  
    }  
}