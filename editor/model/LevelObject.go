package model

import (
	"strconv"

	"github.com/inkyblackness/shocked-model"
)

// LevelObject describes one object within a level
type LevelObject struct {
	data  *model.LevelObject
	index int
}

func newLevelObject(data *model.LevelObject) *LevelObject {
	index, _ := strconv.ParseInt(data.ID, 10, 32)
	obj := &LevelObject{
		data:  data,
		index: int(index)}

	return obj
}

// Index returns the object's index within the level.
func (obj *LevelObject) Index() int {
	return obj.index
}

// ID returns the object ID of the object
func (obj *LevelObject) ID() ObjectID {
	return MakeObjectID(obj.data.Class, obj.data.Subclass, obj.data.Type)
}

// ClassData returns the raw data for the level object.
func (obj *LevelObject) ClassData() []int {
	return obj.data.Hacking.ClassData
}

// Center returns the location of the object within the map
func (obj *LevelObject) Center() (x, y float32) {
	x = (float32(obj.data.BaseProperties.TileX) * 256.0) + float32(obj.data.BaseProperties.FineX)
	y = (float32(63-obj.data.BaseProperties.TileY) * 256.0) + float32(0xFF-obj.data.BaseProperties.FineY)

	return
}
