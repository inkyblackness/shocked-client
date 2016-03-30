package opengl

// OpenGl describes an Open GL interface usable for all environments of this
// application. It is the common subset of WebGL (= OpenGL ES 2) and an equivalent
// API on the desktop.
type OpenGl interface {
	GetError() uint32

	Clear(mask uint32)
	ClearColor(red, green, blue, alpha float32)
}
