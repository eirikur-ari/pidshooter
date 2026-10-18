package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand_Help_ReportsWhetherUsageTextIsAsked(t *testing.T) {
	tests := newCommandHelpTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			command := newCommand(tt.args)

			// When
			help := command.Help()

			// Then
			assert.Equal(t, tt.expectedHelp, help)
		})
	}
}

func TestCommand_Play_ReportsWhetherPlayIsAsked(t *testing.T) {
	tests := newCommandPlayTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			command := newCommand(tt.args)

			// When
			play := command.Play()

			// Then
			assert.Equal(t, tt.expectedPlay, play)
		})
	}
}

func TestCommand_Name_ReturnsFirstArgument(t *testing.T) {
	assert.Equal(t, "chrome", newCommand([]string{"chrome", "play"}).Name())
	assert.Empty(t, newCommand(nil).Name())
}

func TestCommand_Arguments_ReturnsArgumentsAfterName(t *testing.T) {
	assert.Equal(t, []string{"chrome", "--confirm"}, newCommand([]string{"play", "chrome", "--confirm"}).Arguments())
	assert.Empty(t, newCommand(nil).Arguments())
}

func newCommandHelpTestCases() []struct {
	name         string
	args         []string
	expectedHelp bool
} {
	return []struct {
		name         string
		args         []string
		expectedHelp bool
	}{
		{"no arguments", []string{}, true},
		{"help flag", []string{"--help"}, true},
		{"short help flag", []string{"-h"}, true},
		{"help flag after command", []string{"play", "proc", "--help"}, true},
		{"help flag with unknown command", []string{"--unknown", "--help"}, true},
		{"play", strings.Fields("play chrome"), false},
		{"unknown command", []string{"chrome", "play"}, false},
	}
}

func newCommandPlayTestCases() []struct {
	name         string
	args         []string
	expectedPlay bool
} {
	return []struct {
		name         string
		args         []string
		expectedPlay bool
	}{
		{"play", strings.Fields("play chrome"), true},
		{"play without patterns", []string{"play"}, true},
		{"play with help flag", []string{"play", "--help"}, false},
		{"no arguments", []string{}, false},
		{"unknown command", []string{"chrome", "play"}, false},
		{"capitalized command", []string{"Play", "chrome"}, false},
		{"flag before command", []string{"--time=15", "play", "chrome"}, false},
	}
}
