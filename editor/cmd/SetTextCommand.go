package cmd

// SetTextCommand changes a text value.
type SetTextCommand struct {
	Setter   func(value string) error
	OldValue string
	NewValue string
}

// Do sets the new value.
func (cmd SetTextCommand) Do() error {
	return cmd.Setter(cmd.NewValue)
}

// Undo sets the old value.
func (cmd SetTextCommand) Undo() error {
	return cmd.Setter(cmd.OldValue)
}
