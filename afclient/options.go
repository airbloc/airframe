package afclient

type queryOptions struct {
	skip  int
	limit int
}

type QueryOption func(opt *queryOptions)

func WithSkip(skip int) QueryOption {
	return func(opt *queryOptions) {
		opt.skip = skip
	}
}

func WithLimit(limit int) QueryOption {
	return func(opt *queryOptions) {
		opt.limit = limit
	}
}
