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

func (m *MockScreen) Init() error {
	return m.Called().Error(0)
}

func (m *MockScreen) Fini() {
	m.Called()
}
