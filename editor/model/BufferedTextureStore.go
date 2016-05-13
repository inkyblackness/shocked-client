package model

import (
	"github.com/inkyblackness/shocked-client/graphics"
)

// BufferedTextureStore keeps textures in a buffer.
type BufferedTextureStore struct {
	textures map[int]graphics.Texture
}

// NewBufferedTextureStore returns a new instance of a store.
func NewBufferedTextureStore() *BufferedTextureStore {
	return &BufferedTextureStore{
		textures: make(map[int]graphics.Texture)}
}

// Texture returns the texture associated with the given ID. May be null if
// not yet known/available.
func (store *BufferedTextureStore) Texture(id int) graphics.Texture {
	return nil
}

// SetTexture registers a (new) texture under given ID.
func (store *BufferedTextureStore) SetTexture(id int, texture graphics.Texture) {

}
