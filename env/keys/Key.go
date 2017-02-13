package keys

// Key describes a named key on the keyboard. These are keys which are
// unspecific to layout or language, or are universal.
// Or, described in another way: keys that don't end up as printable characters.
type Key int

// Constants for commonly known named keys.
const (
	KeyEnter     = Key(257)
	KeyEscape    = Key(256)
	KeyBackspace = Key(259)
	KeyTab       = Key(258)

	KeyDown  = Key(264)
	KeyLeft  = Key(263)
	KeyRight = Key(262)
	KeyUp    = Key(265)

	KeyDelete   = Key(261)
	KeyEnd      = Key(269)
	KeyHome     = Key(268)
	KeyInsert   = Key(260)
	KeyPageDown = Key(267)
	KeyPageUp   = Key(266)

	KeyAlt     = Key(342)
	KeyControl = Key(341)
	KeyShift   = Key(340)
	KeySuper   = Key(343)

	KeyPause       = Key(284)
	KeyPrintScreen = Key(283)

	KeyCapsLock   = Key(280)
	KeyScrollLock = Key(281)

	KeyF1  = Key(290)
	KeyF10 = Key(299)
	KeyF11 = Key(300)
	KeyF12 = Key(301)
	KeyF13 = Key(302)
	KeyF14 = Key(303)
	KeyF15 = Key(304)
	KeyF16 = Key(305)
	KeyF17 = Key(306)
	KeyF18 = Key(307)
	KeyF19 = Key(308)
	KeyF2  = Key(291)
	KeyF20 = Key(309)
	KeyF21 = Key(310)
	KeyF22 = Key(311)
	KeyF23 = Key(312)
	KeyF24 = Key(313)
	KeyF25 = Key(314)
	KeyF3  = Key(292)
	KeyF4  = Key(293)
	KeyF5  = Key(294)
	KeyF6  = Key(295)
	KeyF7  = Key(296)
	KeyF8  = Key(297)
	KeyF9  = Key(298)
)

var keyToModifier = map[Key]Modifier{
	KeyShift:   ModShift,
	KeyControl: ModControl,
	KeyAlt:     ModAlt,
	KeySuper:   ModSuper}

// AsModifier returns the modifier equivalent for the key - if applicable.
func (key Key) AsModifier() Modifier {
	mod, isModifier := keyToModifier[key]

	if !isModifier {
		mod = ModNone
	}

	return mod
}
