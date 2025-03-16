package goat

import (
	"context"
)

type Component func(ctx context.Context, props any) GoatNode
