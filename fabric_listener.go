package fabric

import (
	"context"
	"reflect"

	"go.uber.org/zap"
)

// Listen on all transports
func (f *Fabric) Listen(ctx context.Context) error {
	// TODO handle re-listening on fail
	// Iterate over available transports and start listening
	for _, t := range f.transports {
		name := reflect.TypeOf(t).String()
		Logger(ctx).Info("Listening on tranport.", zap.String("transport", name))
		if err := t.Listen(ctx, f.Handle); err != nil {
			return err
		}
	}
	return nil
}
