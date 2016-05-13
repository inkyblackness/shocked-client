package model

import (
	check "gopkg.in/check.v1"
)

type BufferedTextureStoreSuite struct {
}

var _ = check.Suite(&BufferedTextureStoreSuite{})

func (suite *BufferedTextureStoreSuite) TestNewStoreReturnsValue(c *check.C) {
	store := NewBufferedTextureStore()

	c.Check(store, check.NotNil)
}
