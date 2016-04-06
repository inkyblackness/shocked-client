package opengl

// Buffer Bits
const (
	DEPTH_BUFFER_BIT   uint32 = 0x00000100
	STENCIL_BUFFER_BIT        = 0x00000400
	COLOR_BUFFER_BIT          = 0x00004000
)

// Draw Types
const (
	POINTS         uint32 = 0x0000
	LINES                 = 0x0001
	LINE_LOOP             = 0x0002
	LINE_STRIP            = 0x0003
	TRIANGLES             = 0x0004
	TRIANGLE_STRIP        = 0x0005
	TRIANGLE_FAN          = 0x0006
)

// Shader Types
const (
	FRAGMENT_SHADER uint32 = 0x8B30
	VERTEX_SHADER          = 0x8B31
)

// Status Values
const (
	COMPILE_STATUS uint32 = 0x8B81
	LINK_STATUS           = 0x8B82
)

// Buffer Types
const (
	ARRAY_BUFFER uint32 = 0x8892
)

// Draw Types
const (
	STREAM_DRAW  uint32 = 0x88E0
	STATIC_DRAW         = 0x88E4
	DYNAMIC_DRAW        = 0x88E8
)

// Features
const (
	DEPTH_TEST uint32 = 0x0B71
)

// Data Types
const (
	BYTE           uint32 = 0x1400
	UNSIGNED_BYTE         = 0x1401
	SHORT                 = 0x1402
	UNSIGNED_SHORT        = 0x1403
	INT                   = 0x1404
	UNSIGNED_INT          = 0x1405
	FLOAT                 = 0x1406
)

// Errors
const (
	NO_ERROR                      uint32 = 0
	INVALID_ENUM                         = 0x0500
	INVALID_VALUE                        = 0x0501
	INVALID_OPERATION                    = 0x0502
	STACK_OVERFLOW                       = 0x0503
	STACK_UNDERFLOW                      = 0x0504
	OUT_OF_MEMORY                        = 0x0505
	INVALID_FRAMEBUFFER_OPERATION        = 0x0506
)

// Color Types
const (
	RGBA uint32 = 0x1908
)
