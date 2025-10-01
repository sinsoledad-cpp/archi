package event

type Consumer interface {
	Start() error
}
