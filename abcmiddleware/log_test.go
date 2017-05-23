package abcmiddleware

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	z, err := zap.NewDevelopment()
	if err != nil {
		t.Error(err)
	}

	ctx = context.WithValue(ctx, CtxLoggerKey, z)
	r := &http.Request{}
	r = r.WithContext(ctx)

	// Ensure log can be called successfully. Ignore response because we don't
	// need to validate anything.
	_ = Log(r)
}
