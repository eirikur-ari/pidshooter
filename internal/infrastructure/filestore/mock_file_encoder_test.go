package filestore

import "github.com/stretchr/testify/mock"

// MockFileEncoder is a test double for fileEncoder.
type MockFileEncoder struct{ mock.Mock }

func (m *MockFileEncoder) encodeYAML(value any) ([]byte, error) {
	return m.encode("encodeYAML", value)
}

func (m *MockFileEncoder) encodeJSON(value any) ([]byte, error) {
	return m.encode("encodeJSON", value)
}

func (m *MockFileEncoder) encode(method string, value any) ([]byte, error) {
	args := m.MethodCalled(method, value)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}
