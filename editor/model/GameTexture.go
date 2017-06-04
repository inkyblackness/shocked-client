package model

import (
	"github.com/inkyblackness/shocked-model"
)

// GameTexture describes one texture available to the game.
type GameTexture struct {
	id int

	properties model.TextureProperties
}

func newGameTexture(id int) *GameTexture {
	return &GameTexture{id: id}
}

// ID uniquely identifies the texture in the game.
func (texture *GameTexture) ID() int {
	return texture.id
}

// Climbable returns whether the texture can be climbed.
func (texture *GameTexture) Climbable() bool {
	return *texture.properties.Climbable
}
