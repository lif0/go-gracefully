package gogracefully

import "context"

// ShutdownAll shuts down all registered GracefulShutdown instances synchronously.
// It uses the provided context for shutdown operations and collects errors in a map,
// keyed by the instance names returned from GracefulShutdownName.
// Returns a map of errors; empty if all shutdowns succeeded.
func (gr *GracefullyRegister) ShutdownAll(ctx context.Context) map[string]error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	errs := make(map[string]error)
	for _, v := range gr.instances {
		instanceErr := v.GracefulShutdown(ctx)
		if instanceErr != nil {
			errs[v.GracefulShutdownName()] = instanceErr
		}
	}

	return errs
}

// func (gr *GracefullyRegister) ShutdownAllAsync() error {
// }
