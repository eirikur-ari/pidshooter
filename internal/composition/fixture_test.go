package composition

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func newRunnerFixture() *Runner {
	return &Runner{errHandler: apperror.NewHandler(&testutil.FakeLogger{})}
}
