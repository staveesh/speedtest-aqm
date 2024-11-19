package channel

type Type = chan struct{}

var (
	IPGrabbed     = make(Type)
	Stop          = make(Type)
	SpeedtestDone = make(Type)
	CaptureDone   = make(Type)
	PingDone      = make(Type)
)
