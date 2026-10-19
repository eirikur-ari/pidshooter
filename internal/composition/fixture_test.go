package composition

import (
	"github.com/eirikur-ari/pidshooter/internal/application/apperror"
	"github.com/eirikur-ari/pidshooter/internal/testutil"
)

func newCreatorFixture() *Creator {
	return &Creator{errHandler: apperror.NewHandler(&testutil.FakeLogger{})}
}
