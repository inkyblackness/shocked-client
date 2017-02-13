package keys

import (
	check "gopkg.in/check.v1"
)

type keyEvent struct {
	down bool
	key  Key
	mod  Modifier
}

type testingStickyKeysListener struct {
	eventMap map[Key][]keyEvent
	events   []keyEvent
}

func (listener *testingStickyKeysListener) KeyDown(key Key, modifier Modifier) {
	listener.addEvent(keyEvent{true, key, modifier})
}

func (listener *testingStickyKeysListener) KeyUp(key Key, modifier Modifier) {
	listener.addEvent(keyEvent{false, key, modifier})
}

func (listener *testingStickyKeysListener) addEvent(event keyEvent) {
	listener.eventMap[event.key] = append(listener.eventMap[event.key], event)
	listener.events = append(listener.events, event)
}

type StickyKeyBufferSuite struct {
	buffer   *StickyKeyBuffer
	listener *testingStickyKeysListener
}

var _ = check.Suite(&StickyKeyBufferSuite{})

func (suite *StickyKeyBufferSuite) SetUpTest(c *check.C) {
	suite.listener = &testingStickyKeysListener{
		eventMap: make(map[Key][]keyEvent)}
	suite.buffer = NewStickyKeyBuffer(suite.listener)
}

func (suite *StickyKeyBufferSuite) TestRegularEventsAreForwarded_A(c *check.C) {
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)
	suite.buffer.KeyDown(KeyF2, ModNone)
	suite.buffer.KeyUp(KeyF2, ModNone)

	c.Check(suite.listener.eventMap[KeyF1], check.DeepEquals, []keyEvent{{true, KeyF1, ModNone}, {false, KeyF1, ModNone}})
	c.Check(suite.listener.eventMap[KeyF2], check.DeepEquals, []keyEvent{{true, KeyF2, ModNone}, {false, KeyF2, ModNone}})
}

func (suite *StickyKeyBufferSuite) TestRegularEventsAreForwarded_B(c *check.C) {
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyDown(KeyF2, ModNone)
	suite.buffer.KeyUp(KeyF2, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)

	c.Check(suite.listener.eventMap[KeyF1], check.DeepEquals, []keyEvent{{true, KeyF1, ModNone}, {false, KeyF1, ModNone}})
	c.Check(suite.listener.eventMap[KeyF2], check.DeepEquals, []keyEvent{{true, KeyF2, ModNone}, {false, KeyF2, ModNone}})
	c.Check(suite.listener.events, check.DeepEquals, []keyEvent{
		{true, KeyF1, ModNone}, {true, KeyF2, ModNone}, {false, KeyF2, ModNone}, {false, KeyF1, ModNone}})
}

func (suite *StickyKeyBufferSuite) TestIdenticalKeysAreReportedOnlyOnce(c *check.C) {
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)

	c.Check(suite.listener.eventMap[KeyF1], check.DeepEquals, []keyEvent{{true, KeyF1, ModNone}, {false, KeyF1, ModNone}})
	c.Check(suite.listener.events, check.DeepEquals, []keyEvent{
		{true, KeyF1, ModNone}, {false, KeyF1, ModNone}})
}

func (suite *StickyKeyBufferSuite) TestSuperfluousReleasesAreIgnored(c *check.C) {
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)
	suite.buffer.KeyDown(KeyF2, ModNone)
	suite.buffer.KeyDown(KeyF1, ModNone)
	suite.buffer.KeyUp(KeyF1, ModNone)

	c.Check(suite.listener.eventMap[KeyF1], check.DeepEquals, []keyEvent{
		{true, KeyF1, ModNone}, {false, KeyF1, ModNone},
		{true, KeyF1, ModNone}, {false, KeyF1, ModNone}})
	c.Check(suite.listener.events, check.DeepEquals, []keyEvent{
		{true, KeyF1, ModNone}, {false, KeyF1, ModNone},
		{true, KeyF2, ModNone},
		{true, KeyF1, ModNone}, {false, KeyF1, ModNone}})
}

func (suite *StickyKeyBufferSuite) TestReleaseAllNotifiesReleasedState(c *check.C) {
	suite.buffer.KeyDown(KeyEnter, ModNone)
	suite.buffer.KeyDown(KeyEnter, ModNone)
	suite.buffer.KeyDown(KeyTab, ModNone)
	suite.buffer.ReleaseAll()

	c.Check(suite.listener.eventMap[KeyEnter], check.DeepEquals, []keyEvent{{true, KeyEnter, ModNone}, {false, KeyEnter, ModNone}})
	c.Check(suite.listener.eventMap[KeyTab], check.DeepEquals, []keyEvent{{true, KeyTab, ModNone}, {false, KeyTab, ModNone}})
}

func (suite *StickyKeyBufferSuite) TestReleaseAllReleasesModifierLast(c *check.C) {
	suite.buffer.KeyDown(KeyShift, ModNone)
	suite.buffer.KeyDown(KeyShift, ModNone)
	suite.buffer.KeyDown(KeyTab, ModShift)
	suite.buffer.ReleaseAll()

	c.Check(suite.listener.events, check.DeepEquals, []keyEvent{
		{true, KeyShift, ModNone}, {true, KeyTab, ModShift},
		{false, KeyTab, ModShift},
		{false, KeyShift, ModShift}})
}

func (suite *StickyKeyBufferSuite) TestActiveModifierReturnsCurrentModifier(c *check.C) {
	suite.buffer.KeyDown(KeyShift, ModNone)
	suite.buffer.KeyDown(KeyControl, ModShift)
	suite.buffer.KeyDown(KeyAlt, ModShift.With(ModControl))
	suite.buffer.KeyUp(KeyControl, ModShift.With(ModControl).With(ModAlt))

	c.Check(suite.buffer.ActiveModifier(), check.Equals, ModShift.With(ModAlt))
}
