package lifecycle

type Lifecycle interface {
	Start() error
	Stop() error
}
