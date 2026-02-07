package recorder

// Recorder captures audio from the default input device.
type Recorder interface {
	Start() error
	Stop() ([]int16, error)
	IsRecording() bool
}
