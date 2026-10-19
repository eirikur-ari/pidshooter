package cli

import (
	"bytes"

	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

type programFixture struct {
	program *Program
	play    *MockPlayRunner
	list    *MockListRunner
	kill    *MockKillRunner
	out     *bytes.Buffer
	errOut  *bytes.Buffer
	logger  *testutil.FakeLogger
}

func newProgramFixture() programFixture {
	fixture := programFixture{
		play:   &MockPlayRunner{},
		list:   &MockListRunner{},
		kill:   &MockKillRunner{},
		out:    &bytes.Buffer{},
		errOut: &bytes.Buffer{},
		logger: &testutil.FakeLogger{},
	}
	fixture.program = NewProgram(fixture.play, fixture.list, fixture.kill, apperror.NewHandler(fixture.logger))
	fixture.program.out, fixture.program.errOut = fixture.out, fixture.errOut
	return fixture
}
