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

// TransparencyControl returns how the pixel data shall be interpreted.
func (texture *GameTexture) TransparencyControl() int {
	return *texture.properties.TransparencyControl
}

// AnimationGroup associates textures of an animation.
func (texture *GameTexture) AnimationGroup() int {
	return *texture.properties.AnimationGroup
}

// AnimationIndex places the texture within the group.
func (texture *GameTexture) AnimationIndex() int {
	return *texture.properties.AnimationIndex
}
