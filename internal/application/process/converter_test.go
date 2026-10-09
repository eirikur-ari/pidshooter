package process

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func TestToInfos_MapsEveryElementInOrder(t *testing.T) {
	tests := newToInfosTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toInfos(test.infos)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToProcessInfos_MapsEveryElementInOrder(t *testing.T) {
	tests := newToProcessInfosTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toProcessInfos(test.infos)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToFindResults_MapsEveryElementInOrder(t *testing.T) {
	tests := newToFindResultsTestCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			actual := toFindResults(test.infos)

			// Then
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newToInfosTestCases() []struct {
	name     string
	infos    []outbound.ProcessInfo
	expected []process.Info
} {
	return []struct {
		name     string
		infos    []outbound.ProcessInfo
		expected []process.Info
	}{
		{
			name: "several infos",
			infos: []outbound.ProcessInfo{
				{PID: 1, Name: "a", Rss: 100, UID: 1000},
				{PID: 2, Name: "b", Rss: 200, UID: 2000},
			},
			expected: []process.Info{
				process.NewInfo(1, "a", 100, 1000),
				process.NewInfo(2, "b", 200, 2000),
			},
		},
		{name: "no infos", infos: nil, expected: []process.Info{}},
	}
}

func newToProcessInfosTestCases() []struct {
	name     string
	infos    []process.Info
	expected []outbound.ProcessInfo
} {
	return []struct {
		name     string
		infos    []process.Info
		expected []outbound.ProcessInfo
	}{
		{
			name: "several infos",
			infos: []process.Info{
				process.NewInfo(1, "a", 100, 1000),
				process.NewInfo(2, "b", 200, 2000),
			},
			expected: []outbound.ProcessInfo{
				{PID: 1, Name: "a", Rss: 100, UID: 1000},
				{PID: 2, Name: "b", Rss: 200, UID: 2000},
			},
		},
		{name: "no infos", infos: nil, expected: []outbound.ProcessInfo{}},
	}
}

func newToFindResultsTestCases() []struct {
	name     string
	infos    []process.Info
	expected []FindResult
} {
	return []struct {
		name     string
		infos    []process.Info
		expected []FindResult
	}{
		{
			name: "several infos",
			infos: []process.Info{
				process.NewInfo(1, "a", 100, 1000),
				process.NewInfo(2, "b", 200, 2000),
			},
			expected: []FindResult{
				{PID: 1, Name: "a", Rss: 100, UID: 1000},
				{PID: 2, Name: "b", Rss: 200, UID: 2000},
			},
		},
		{name: "no infos", infos: nil, expected: []FindResult{}},
	}
}
