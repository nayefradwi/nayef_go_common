package connectutil

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"connectrpc.com/connect"
)

type RecoveryOptions struct {
	Logger *slog.Logger
}

func Recover(opts ...RecoveryOptions) func(context.Context, connect.Spec, http.Header, any) error {
	opt := RecoveryOptions{Logger: slog.Default()}
	if len(opts) > 0 {
		opt = opts[0]
	}

	return func(_ context.Context, spec connect.Spec, _ http.Header, recovered any) error {
		err, ok := recovered.(error)
		if !ok {
			err = fmt.Errorf("%v", recovered)
		}

		opt.Logger.Error("panic recovered in connect handler",
			"procedure", spec.Procedure,
			"error", err,
			"stack", string(debug.Stack()),
		)

		return toServerError(err)
	}
}
