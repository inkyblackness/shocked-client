package model

import (
	"fmt"

	"github.com/inkyblackness/shocked-model"
)

// Adapter is the central model adapter.
type Adapter struct {
	store model.DataStore

	message *observable

	activeProjectID     *observable
	availableArchiveIDs *observable
	activeArchiveID     *observable
	activeLevel         *LevelAdapter

	availableLevels   map[string]model.LevelProperties
	availableLevelIDs *observable
}

// NewAdapter returns a new model adapter.
func NewAdapter(store model.DataStore) *Adapter {
	adapter := &Adapter{
		store:   store,
		message: newObservable(),

		activeProjectID:     newObservable(),
		availableArchiveIDs: newObservable(),
		activeArchiveID:     newObservable(),
		activeLevel:         newLevelAdapter(),

		availableLevels:   make(map[string]model.LevelProperties),
		availableLevelIDs: newObservable()}

	adapter.message.set("")

	return adapter
}

func (adapter *Adapter) simpleStoreFailure(info string) model.FailureFunc {
	return func() {
		adapter.SetMessage(fmt.Sprintf("Failed to process store query <%s>", info))
	}
}

// SetMessage sets the current global message.
func (adapter *Adapter) SetMessage(message string) {
	adapter.message.set(message)
}

// Message returns the current global message.
func (adapter *Adapter) Message() string {
	return adapter.message.get().(string)
}

// OnMessageChanged registers a callback for the global message.
func (adapter *Adapter) OnMessageChanged(callback func()) {
	adapter.message.addObserver(callback)
}

// ActiveProjectID returns the identifier of the current project.
func (adapter *Adapter) ActiveProjectID() string {
	return adapter.activeProjectID.get().(string)
}

// RequestProject sets the project to work on.
func (adapter *Adapter) RequestProject(projectID string) {
	adapter.requestArchive("")
	adapter.availableArchiveIDs.set("")

	adapter.activeProjectID.set(projectID)
	if projectID != "" {
		adapter.availableArchiveIDs.set([]string{"archive"})
		adapter.requestArchive("archive")
	}
}

// ActiveArchiveID returns the identifier of the current archive.
func (adapter *Adapter) ActiveArchiveID() string {
	return adapter.activeArchiveID.get().(string)
}

func (adapter *Adapter) requestArchive(archiveID string) {
	adapter.RequestActiveLevel("")
	adapter.availableLevels = make(map[string]model.LevelProperties)
	adapter.availableLevelIDs.set(nil)

	adapter.activeArchiveID.set(archiveID)
	if archiveID != "" {
		adapter.store.Levels(adapter.ActiveProjectID(), adapter.ActiveArchiveID(),
			adapter.onLevels,
			adapter.simpleStoreFailure("Levels"))
	}
}

func (adapter *Adapter) onLevels(levels []model.Level) {
	availableLevelIDs := make([]string, len(levels))

	adapter.availableLevels = make(map[string]model.LevelProperties)
	for index, entry := range levels {
		availableLevelIDs[index] = entry.ID
		adapter.availableLevels[entry.ID] = entry.Properties
	}
	adapter.availableLevelIDs.set(availableLevelIDs)
}

// RequestActiveLevel requests to set the specified level as the active one.
func (adapter *Adapter) RequestActiveLevel(levelID string) {
	adapter.activeLevel.clear(levelID)
	// clear current level stuff
	// set active level (callback)
	// request all level specific stuff
}

// AvailableLevelIDs returns the list of identifier of available levels.
func (adapter *Adapter) AvailableLevelIDs() []string {
	return adapter.availableLevelIDs.get().([]string)
}

// OnAvailableLevelsChanged registers a callback for changes of available levels.
func (adapter *Adapter) OnAvailableLevelsChanged(callback func()) {
	adapter.availableLevelIDs.addObserver(callback)
}

// introduce Observable type with change callback
// alt: use bus?
//
