package orm

// Default config
func DefaultConfig() *Conf {
	return &Conf{
		DebugLogging:    true,
		SafeModeEnabled: true,
		AutoOpen:        true,
		Dialect:         "mysql",
	}
}

// Logging
type Conf struct {
	DebugLogging    bool
	SafeModeEnabled bool
	AutoOpen        bool

	Username         string
	Password         string
	Hostname         string
	Port             int
	Database         string
	Dialect          string
	ConnectionString string // override
}
