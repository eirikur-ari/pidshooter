package tcellui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/mock"
)

type MockScreen struct {
	tcell.Screen
	mock.Mock
}

func (m *MockScreen) Sync() {
	m.Called()
}
