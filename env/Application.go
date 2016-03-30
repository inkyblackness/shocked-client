package env

// Application represents the public interface between the environment and the
// actual application core.
type Application interface {
	Init(window OpenGlWindow)
}
