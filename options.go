package migrationsql

type options struct {
	source SourceWithAdd
}

type Option func(*options)

// WithSource allows to set a custom source to be used by the migrations.
func WithSource(source SourceWithAdd) Option {
	return func(o *options) {
		o.source = source
	}
}
