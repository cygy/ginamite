package functions

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// DeleteExpired deletes the expired auth tokens and returns the number of tokens deleted.
func (f *Functions) DeleteExpired(ctx context.Context) (int64, error) {
	count, err := model.DeleteExpiredAuthTokens(ctx, f.database())
	if err != nil {
		return count, fmt.Errorf("unable to delete the expired auth tokens: %w", err)
	}

	return count, nil
}
