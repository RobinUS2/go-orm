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
}
