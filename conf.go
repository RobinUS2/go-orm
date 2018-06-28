package orm

// Default config
func DefaultConfig() *Conf {
	return &Conf{
		DebugLogging:    true,
		SafeModeEnabled: true,
	}
}

// Logging
type Conf struct {
	DebugLogging    bool
	SafeModeEnabled bool

	Username string
	Password string
	Hostname string
	Port int
	Database string
}
