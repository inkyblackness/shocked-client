package keys

// StickyKeyListener is the listener interface for receiving key events.
type StickyKeyListener interface {
	KeyDown(key Key, modifier Modifier)
	KeyUp(key Key, modifier Modifier)
}

// StickyKeyBuffer is a buffer to keep track of several identically named keys.
// Keys can be reported being pressed or released. Their state will be forwarded
// to a StickyKeyListener instance. If a specific key is reported to be pressed
// more than once, the listener will have received the down state only once.
// Only after the key has been released the equal number of times, the listener
// will received the release event.
type StickyKeyBuffer struct {
	pressedKeys    map[Key]int
	activeModifier Modifier
	listener       StickyKeyListener
}

// NewStickyKeyBuffer returns a new instance of a sticky key buffer.
func NewStickyKeyBuffer(listener StickyKeyListener) *StickyKeyBuffer {
	buffer := &StickyKeyBuffer{
		pressedKeys:    make(map[Key]int),
		activeModifier: ModNone,
		listener:       listener}

	return buffer
}

// ActiveModifier returns the currently pressed modifier set.
func (buffer *StickyKeyBuffer) ActiveModifier() Modifier {
	return buffer.activeModifier
}

// KeyDown registers a pressed key state. Multiple down states can be
// registered for the same key.
func (buffer *StickyKeyBuffer) KeyDown(key Key, modifier Modifier) {
	oldCount := buffer.pressedKeys[key]

	buffer.pressedKeys[key] = oldCount + 1
	if oldCount == 0 {
		buffer.activeModifier = buffer.activeModifier.With(key.AsModifier())
		buffer.listener.KeyDown(key, modifier)
	}
}

// KeyUp registers a released key state. Multiple up states can be registered
// for the same key, as long as enough down states were reported.
func (buffer *StickyKeyBuffer) KeyUp(key Key, modifier Modifier) {
	oldCount := buffer.pressedKeys[key]

	if oldCount > 0 {
		buffer.pressedKeys[key] = oldCount - 1
		if oldCount == 1 {
			buffer.activeModifier = buffer.activeModifier.Without(key.AsModifier())
			buffer.listener.KeyUp(key, modifier)
		}
	}
}

// ReleaseAll notifies the listener of up states of all currently pressed
// keys. Non-modifier keys are reported with the currently active modifiers,
// and modifier keys themselves are reported last.
func (buffer *StickyKeyBuffer) ReleaseAll() {
	var pendingModifierKeys []Key

	for key, count := range buffer.pressedKeys {
		if count > 0 {
			modifier := key.AsModifier()

			if modifier == ModNone {
				buffer.listener.KeyUp(key, buffer.activeModifier)
			} else {
				pendingModifierKeys = append(pendingModifierKeys, key)
			}
		}
	}
	for _, key := range pendingModifierKeys {
		buffer.listener.KeyUp(key, buffer.activeModifier)
		buffer.activeModifier = buffer.activeModifier.Without(key.AsModifier())
	}

	buffer.pressedKeys = make(map[Key]int)
	buffer.activeModifier = ModNone
}
