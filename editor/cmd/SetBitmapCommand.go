package cmd

import "github.com/inkyblackness/shocked-model"

// SetBitmapCommand changes an audio clip.
type SetBitmapCommand struct {
	Setter   func(key model.ResourceKey, bmp *model.RawBitmap) error
	Key      model.ResourceKey
	OldValue *model.RawBitmap
	NewValue *model.RawBitmap
}

// Do sets the new value.
func (cmd SetBitmapCommand) Do() error {
	return cmd.Setter(cmd.Key, cmd.NewValue)
}

// Undo sets the old value.
func (cmd SetBitmapCommand) Undo() error {
	return cmd.Setter(cmd.Key, cmd.OldValue)
}
