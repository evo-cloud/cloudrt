package jobs

// Revertable builds a TaskFn splitting rollback from normal logic
func Revertable(fwd, rev TaskFn) TaskFn {
	return func(ctx Context) error {
		if !ctx.IsRollback() {
			return fwd(ctx)
		} else if rev != nil {
			return rev(ctx)
		}
		return nil
	}
}

// NonRevertable builds a TaskFn fails on rollback
func NonRevertable(fn TaskFn) TaskFn {
	return func(ctx Context) error {
		if ctx.IsRollback() {
			return ctx.Fail(ErrTaskNonRevertable)
		}
		return fn(ctx)
	}
}
