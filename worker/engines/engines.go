package engines

// EngineImpl describes the functions a password cracking engine must export
type EngineImpl interface {
	// Initialize the engine; allocating any resource needed to start
	Initialize() error
	// Start cracking the task. This should block
	Start() error
	// Stop the engine
	Stop() error
	// GetStatus should return the status of the password cracking task. e.g. estimated time to completion, progress, etc
	GetStatus() interface{}
	// Cleanup is called after the engine stops and should release all resources
	Cleanup()
}
