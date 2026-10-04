package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfirmation_Pending_IsFalseByDefault(t *testing.T) {
	// Given
	confirm := confirmation{}

	// When
	result := confirm.Pending()

	// Then
	assert.False(t, result)
}

func TestConfirmation_Pending_IsTrueWhenTargetIsSet(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture()}
	confirm := confirmation{target: target}

	// When
	result := confirm.Pending()

	// Then
	assert.True(t, result)
}

func TestConfirmation_Request_SetsTargetAndReturnsNilWhenConfirmIsTrue(t *testing.T) {
	// Given
	confirm := newConfirmation(true)
	target := &Target{Info: newInfoFixture()}

	// When
	result := confirm.Request(target)

	// Then
	assert.Nil(t, result)
	assert.NotNil(t, confirm.target)
}

func TestConfirmation_Request_ReturnsTargetWhenConfirmIsFalse(t *testing.T) {
	// Given
	confirm := newConfirmation(false)
	target := &Target{Info: newInfoFixture()}

	// When
	result := confirm.Request(target)

	// Then
	assert.Equal(t, target, result)
	assert.Nil(t, confirm.target)
}

func TestConfirmation_Accept_ReturnsNilByDefault(t *testing.T) {
	// Given
	confirm := newConfirmation(false)

	// When
	result := confirm.Accept()

	// Then
	assert.Nil(t, result)
	assert.Nil(t, confirm.target)
}

func TestConfirmation_Accept_ReturnsTargetAndClearsTargetField(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture()}
	confirm := confirmation{target: target}

	// When
	result := confirm.Accept()

	// Then
	assert.Equal(t, target, result)
	assert.Nil(t, confirm.target)
}

func TestConfirmation_Cancel_ClearsTargetField(t *testing.T) {
	// Given
	target := &Target{Info: newInfoFixture()}
	confirm := confirmation{target: target}

	// When
	confirm.Cancel()

	// Then
	assert.Nil(t, confirm.target)
}
