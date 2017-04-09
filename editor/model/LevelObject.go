package model

import (
	"strconv"

	"github.com/inkyblackness/shocked-model"
)

// LevelObject describes one object within a level
type LevelObject struct {
	class            int
	properties       *model.LevelObjectProperties
	centerX, centerY float32
	index            int
}

func newLevelObject(data *model.LevelObject) *LevelObject {
	index, _ := strconv.ParseInt(data.ID, 10, 32)
	obj := &LevelObject{
		class: data.Class,
		index: int(index)}
	obj.onPropertiesChanged(&data.Properties)

	return obj
}

func (obj *LevelObject) onPropertiesChanged(properties *model.LevelObjectProperties) {
	obj.properties = properties
	obj.centerX = float32((*obj.properties.TileX << 8) + *obj.properties.FineX)
	obj.centerY = float32((*obj.properties.TileY << 8) + *obj.properties.FineY)
}

// Index returns the object's index within the level.
func (obj *LevelObject) Index() int {
	return obj.index
}

// ID returns the object ID of the object
func (obj *LevelObject) ID() ObjectID {
	return MakeObjectID(obj.class, *obj.properties.Subclass, *obj.properties.Type)
}

// ClassData returns the raw data for the level object.
func (obj *LevelObject) ClassData() []byte {
	return obj.properties.ClassData
}

// Z returns the z-coordinate (placement height) of the object
func (obj *LevelObject) Z() int {
	return *obj.properties.Z
}

// Center returns the location of the object within the map
func (obj *LevelObject) Center() (x, y float32) {
	return obj.centerX, obj.centerY
}
