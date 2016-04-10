package editor

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/inkyblackness/shocked-model"
)

// RestDataStore is a REST based implementation of the DataStore interface.
type RestDataStore struct {
	transport RestTransport
}

// NewRestDataStore returns a new instance of a data store backed by a REST transport.
func NewRestDataStore(transport RestTransport) *RestDataStore {
	return &RestDataStore{transport: transport}
}

func (store *RestDataStore) get(url string, responseData interface{}, onSuccess func(), onFailure FailureFunc) {
	store.transport.Get(url, func(jsonString string) {
		json.Unmarshal(bytes.NewBufferString(jsonString).Bytes(), responseData)
		onSuccess()
	}, func() {
		onFailure()
	})
}

// Palette implements the DataStore interface.
func (store *RestDataStore) Palette(projectID string, paletteID string,
	onSuccess func(colors [256]model.Color), onFailure FailureFunc) {
	url := fmt.Sprintf("/projects/%s/palettes/%s", projectID, paletteID)
	var data model.Palette

	store.get(url, &data, func() {
		onSuccess(data.Colors)
	}, onFailure)
}

// LevelTextures implements the DataStore interface.
func (store *RestDataStore) LevelTextures(projectID string, archiveID string, levelID int,
	onSuccess func(textureIDs []int), onFailure FailureFunc) {
	url := fmt.Sprintf("/projects/%s/%s/levels/%d/textures", projectID, archiveID, levelID)
	var data model.LevelTextures

	store.get(url, &data, func() {
		onSuccess(data.IDs)
	}, onFailure)
}

// TextureBitmap implements the DataStore interface.
func (store *RestDataStore) TextureBitmap(projectID string, textureID int, size string,
	onSuccess func(bmp *model.RawBitmap), onFailure FailureFunc) {
	url := fmt.Sprintf("/projects/%s/textures/%d/%s/raw", projectID, textureID, size)
	var data model.RawBitmap

	store.get(url, &data, func() {
		onSuccess(&data)
	}, onFailure)
}

// Tiles implements the DataStore interface.
func (store *RestDataStore) Tiles(projectID string, archiveID string, levelID int,
	onSuccess func(tiles model.Tiles), onFailure FailureFunc) {
	url := fmt.Sprintf("/projects/%s/%s/levels/%d/tiles", projectID, archiveID, levelID)
	var data model.Tiles

	store.get(url, &data, func() {
		onSuccess(data)
	}, onFailure)
}
