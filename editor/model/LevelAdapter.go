package model

// LevelAdapter is the entry point for a level.
type LevelAdapter struct {
	id *observable
}

func newLevelAdapter() *LevelAdapter {
	adapter := &LevelAdapter{
		id: newObservable()}

	adapter.id.set("")

	return adapter
}

func (adapter *LevelAdapter) clear(levelID string) {
	adapter.id.set("")

	adapter.id.set(levelID)
}

// ID returns the ID of the level.
func (adapter *LevelAdapter) ID() string {
	return adapter.id.orDefault("").(string)
}

// OnIDChanged registers a callback for changed IDs.
func (adapter *LevelAdapter) OnIDChanged(callback func()) {
	adapter.id.addObserver(callback)
}
