package strategies

import "context"

type Strategy interface {
	Run(ctx context.Context) error
}
