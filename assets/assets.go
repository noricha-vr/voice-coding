package assets

import _ "embed"

//go:embed icon_idle.png
var IconIdle []byte

//go:embed icon_recording.png
var IconRecording []byte

//go:embed icon_processing.png
var IconProcessing []byte

//go:embed sounds/start.wav
var SoundStart []byte

//go:embed sounds/stop.wav
var SoundStop []byte

//go:embed sounds/success.wav
var SoundSuccess []byte

//go:embed sounds/error.wav
var SoundError []byte
