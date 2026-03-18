type RecoveryOptions struct {
	Logger *slog.Logger
}

func RecoverUnary(opts ...RecoveryOptions) grpc.UnaryServerInterceptor {
	opt := RecoveryOptions{Logger: slog.Default()}
	if len(opts) > 0 {
		opt = opts[0]
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handleRecoveredPanic(r, info.FullMethod, opt.Logger)
			}
		}()
		return handler(ctx, req)
	}
}

func handleRecoveredPanic(r any, method string, log *slog.Logger) error {
	var err error
	if e, ok := r.(error); ok {
		err = e
	} else {
		err = fmt.Errorf("%v", r)
	}

	log.Error("panic recovered in gRPC handler",
		"method", method,
		"error", err,
		"stack", string(debug.Stack()),
	)

	return NewGrpcResponseWriter[any]().WriteError(err)
}
