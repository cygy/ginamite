// Package reporter plays the code of a database plugin for the tests of the log package: it reports an error through
// a wrapper, as the functions of the document package of the MongoDB plugin do.
package reporter

import (
	"context"

	"github.com/cygy/ginamite/common/log"
)

// Report reports the error of the operation Aggregate on the collection scores.
func Report(ctx context.Context, err error) { reportWrapped(ctx, err) }

func reportWrapped(ctx context.Context, err error) {
	log.DatabaseOperationError(ctx, err, "Aggregate", "scores")
}
