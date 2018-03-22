package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestCommand struct {
	name string

	executed int
	reverted int

	pendingError error
	task         func()
}

func (cmd *TestCommand) Do() error {
	cmd.executed++
	return cmd.run()
}

func (cmd *TestCommand) Undo() error {
	cmd.reverted++
	return cmd.run()
}

func (cmd *TestCommand) run() (err error) {
	if cmd.task != nil {
		cmd.task()
	}
	err = cmd.pendingError
	cmd.pendingError = nil
	return
}

type StackSuite struct {
	suite.Suite

	stack    *Stack
	commands map[string]*TestCommand
}

func TestStackSuite(t *testing.T) {
	suite.Run(t, new(StackSuite))
}

func (suite *StackSuite) SetupTest() {
	suite.stack = nil
	suite.commands = make(map[string]*TestCommand)
}

func (suite *StackSuite) TestNewStackCantDoAnything() {
	suite.whenAnInstanceIsCreated()
	suite.thenStackShouldNotSupportRedo()
	suite.thenStackShouldNotSupportUndo()
}

func (suite *StackSuite) TestPerformExecutesCommand() {
	suite.givenAnInstance()
	suite.whenPerforming(suite.aCommand("cmd1"))
	suite.thenCommandShouldHaveBeenExecuted("cmd1")
}

func (suite *StackSuite) TestPerformAllowsUndoIfSuccessful() {
	suite.givenAnInstance()
	suite.whenPerforming(suite.aCommand("cmd1"))
	suite.thenStackShouldSupportUndo()
}

func (suite *StackSuite) TestPerformIgnoresCommandIfItFails() {
	suite.givenAnInstance()
	suite.whenPerforming(suite.aCommandReturningError())
	suite.thenStackShouldNotSupportUndo()
}

func (suite *StackSuite) TestPerformReturnsErrorOfCommand() {
	err := fmt.Errorf("fail first time")
	suite.whenAnInstanceIsCreated()
	suite.thenPerformShouldReturnError(suite.aCommandReturning(err), err)
}

func (suite *StackSuite) TestUndoRevertsCommand() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.whenUndoing()
	suite.thenCommandShouldHaveBeenReverted("cmd1")
}

func (suite *StackSuite) TestUndoRevertsCommandOnlyOnce() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenUndoWasCalledTimes(1)
	suite.whenUndoing()
	suite.thenCommandShouldHaveBeenRevertedTimes("cmd1", 1)
}

func (suite *StackSuite) TestUndoRevertsCommandsInSequence() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenCommandWasPerformed("cmd2")

	suite.whenUndoing()
	suite.thenCommandShouldHaveBeenReverted("cmd2")

	suite.whenUndoing()
	suite.thenCommandShouldHaveBeenReverted("cmd1")
}

func (suite *StackSuite) TestUndoLeavesStackUnchangedIfCommandFails() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenCommandWasPerformed("cmd2")
	suite.givenCommandWillFail("cmd2")
	suite.givenUndoWasCalledTimes(1)
	suite.whenUndoing()
	suite.thenCommandShouldHaveBeenRevertedTimes("cmd2", 2)
}

func (suite *StackSuite) TestUndoEnablesRedo() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.whenUndoing()
	suite.thenStackShouldSupportRedo()
}

func (suite *StackSuite) TestRedoExecutesCommandAgain() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenUndoWasCalledTimes(1)
	suite.whenRedoing()
	suite.thenCommandShouldHaveBeenExecutedTimes("cmd1", 2)
}

func (suite *StackSuite) TestRedoExecutesCommandOnlyOnce() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenUndoWasCalledTimes(1)
	suite.givenRedoWasCalledTimes(1)
	suite.whenRedoing()
	suite.thenCommandShouldHaveBeenExecutedTimes("cmd1", 2)
}

func (suite *StackSuite) TestRedoExecutesCommandsInSequence() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenCommandWasPerformed("cmd2")
	suite.givenUndoWasCalledTimes(2)

	suite.whenRedoing()
	suite.thenCommandShouldHaveBeenExecutedTimes("cmd1", 2)

	suite.whenRedoing()
	suite.thenCommandShouldHaveBeenExecutedTimes("cmd2", 2)
}

func (suite *StackSuite) TestRedoLeavesStackUnchangedIfCommandFails() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenUndoWasCalledTimes(1)
	suite.givenCommandWillFail("cmd1")
	suite.givenRedoWasCalledTimes(1)
	suite.whenRedoing()
	suite.thenCommandShouldHaveBeenExecutedTimes("cmd1", 3)
}

func (suite *StackSuite) TestRedoMakesCommandsUndoableAgain() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenUndoWasCalledTimes(1)
	suite.givenRedoWasCalledTimes(1)
	suite.whenUndoing()
	suite.thenCommandShouldHaveBeenRevertedTimes("cmd1", 2)
}

func (suite *StackSuite) TestPerformDropsPendingRedoStack() {
	suite.givenAnInstance()
	suite.givenCommandWasPerformed("cmd1")
	suite.givenUndoWasCalledTimes(1)

	suite.whenPerforming(suite.aCommand("cmd2"))
	suite.thenStackShouldNotSupportRedo()

	suite.whenRedoing()
	suite.thenCommandShouldHaveBeenExecutedTimes("cmd1", 1)
}

func (suite *StackSuite) TestPerformPanicsIfStackIsInUse() {
	callPerform := func(name string) func() {
		var times int
		return func() {
			require.Equal(suite.T(), 0, times, "Re-entrant call detected of <"+name+">. Should actually panic.")
			times++

			assert.Panics(suite.T(), func() {
				suite.stack.Perform(suite.aCommand(name + "-nested"))
			}, "Perform during function <"+name+"> should panic")
		}
	}

	suite.assertPanics(callPerform)
}

func (suite *StackSuite) TestUndoPanicsIfStackIsInUse() {
	callUndo := func(name string) func() {
		var times int
		return func() {
			require.Equal(suite.T(), 0, times, "Re-entrant call detected of <"+name+">. Should actually panic.")
			times++

			assert.Panics(suite.T(), func() {
				suite.stack.Undo()
			}, "Undo during function <"+name+"> should panic")
		}
	}

	suite.assertPanics(callUndo)
}

func (suite *StackSuite) TestRedoPanicsIfStackIsInUse() {
	callRedo := func(name string) func() {
		var times int
		return func() {
			require.Equal(suite.T(), 0, times, "Re-entrant call detected of <"+name+">. Should actually panic.")
			times++

			assert.Panics(suite.T(), func() {
				suite.stack.Undo()
			}, "Redo during function <"+name+"> should panic")
		}
	}

	suite.assertPanics(callRedo)
}

func (suite *StackSuite) assertPanics(taskFor func(string) func()) {
	suite.givenAnInstance()

	cmd1 := suite.aCommandExecuting("cmd1", taskFor("Perform"))
	suite.whenPerforming(cmd1)

	suite.givenCommandWasPerformed("cmd2")
	suite.givenCommandExecutes("cmd2", taskFor("Undo"))
	suite.whenUndoing()

	suite.givenCommandWasPerformed("cmd3")
	suite.givenUndoWasCalledTimes(1)
	suite.givenCommandExecutes("cmd3", taskFor("Redo"))
	suite.whenRedoing()
}

func (suite *StackSuite) givenAnInstance() {
	suite.whenAnInstanceIsCreated()
}

func (suite *StackSuite) givenCommandWasPerformed(name string) {
	suite.whenPerforming(suite.aCommand(name))
}

func (suite *StackSuite) givenUndoWasCalledTimes(times int) {
	for i := 0; i < times; i++ {
		suite.stack.Undo()
	}
}

func (suite *StackSuite) givenRedoWasCalledTimes(times int) {
	for i := 0; i < times; i++ {
		suite.stack.Redo()
	}
}

func (suite *StackSuite) givenCommandWillFail(name string) {
	suite.pastCommand(name).pendingError = fmt.Errorf("failing")
}

func (suite *StackSuite) givenCommandExecutes(name string, task func()) {
	cmd := suite.pastCommand(name)
	cmd.task = task
}

func (suite *StackSuite) whenUndoing() {
	suite.stack.Undo()
}

func (suite *StackSuite) whenRedoing() {
	suite.stack.Redo()
}

func (suite *StackSuite) whenAnInstanceIsCreated() {
	suite.stack = NewStack()
}

func (suite *StackSuite) whenPerforming(command Command) {
	suite.stack.Perform(command)
}

func (suite *StackSuite) thenStackShouldSupportRedo() {
	assert.True(suite.T(), suite.stack.CanRedo(), "Stack should be able to redo")
}

func (suite *StackSuite) thenStackShouldNotSupportRedo() {
	assert.False(suite.T(), suite.stack.CanRedo(), "Stack should not be able to redo")
}

func (suite *StackSuite) thenStackShouldSupportUndo() {
	assert.True(suite.T(), suite.stack.CanUndo(), "Stack should be able to undo")
}

func (suite *StackSuite) thenStackShouldNotSupportUndo() {
	assert.False(suite.T(), suite.stack.CanUndo(), "Stack should not be able to undo")
}

func (suite *StackSuite) thenCommandShouldHaveBeenExecuted(name string) {
	cmd := suite.pastCommand(name)
	assert.True(suite.T(), cmd.executed > 0, "Command <"+name+"> should have been executed at least once")
}

func (suite *StackSuite) thenCommandShouldHaveBeenExecutedTimes(name string, expected int) {
	cmd := suite.pastCommand(name)
	assert.Equal(suite.T(), expected, cmd.executed)
}

func (suite *StackSuite) thenCommandShouldHaveBeenReverted(name string) {
	cmd := suite.pastCommand(name)
	assert.True(suite.T(), cmd.reverted > 0, "Command <"+name+"> should have been reverted at least once")
}

func (suite *StackSuite) thenCommandShouldHaveBeenRevertedTimes(name string, expected int) {
	cmd := suite.pastCommand(name)
	assert.Equal(suite.T(), expected, cmd.reverted)
}

func (suite *StackSuite) thenPerformShouldReturnError(cmd Command, expected error) {
	result := suite.stack.Perform(cmd)
	assert.Equal(suite.T(), expected, result)
}

func (suite *StackSuite) aCommand(name string) Command {
	cmd := &TestCommand{name: name}
	suite.commands[name] = cmd
	return cmd
}

func (suite *StackSuite) aCommandExecuting(name string, task func()) Command {
	cmd := &TestCommand{name: name, task: task}
	suite.commands[name] = cmd
	return cmd
}

func (suite *StackSuite) aCommandReturningError() Command {
	return suite.aCommandReturning(fmt.Errorf("fail"))
}

func (suite *StackSuite) aCommandReturning(err error) Command {
	name := "unnamed"
	cmd := &TestCommand{name: name, pendingError: err}
	suite.commands[name] = cmd
	return cmd
}

func (suite *StackSuite) pastCommand(name string) *TestCommand {
	cmd, found := suite.commands[name]
	if !found {
		require.True(suite.T(), found, "Command not found <"+name+"> - test is wrong")
	}
	return cmd
}
