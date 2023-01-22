package contextutil

import (
	"context"
	"errors"
)

var ErrIncorrectNumberOfItems = errors.New("number of items is not even")

// WithValues combines multiple key-value into an existing context
func WithValues(ctx context.Context, keyValue ...string) (context.Context, error) {
	if len(keyValue)%2 != 0 {
		return ctx, ErrIncorrectNumberOfItems
	}

	for i := 0; i < len(keyValue)-1; i++ {
		ctx = context.WithValue(ctx, keyValue[i], keyValue[i+1]) //nolint
	}
	return ctx, nil
}
