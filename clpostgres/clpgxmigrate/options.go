package clpgxmigrate

import (
	"log/slog"
)

type Option func(*opts)

func VersionTableName(s string) Option {
	return func(o *opts) {
		o.versionTableName = s
	}
}

func VersionSchemaName(s string) Option {
	return func(o *opts) {
		o.versionSchemaName = s
	}
}

func WithCollection(c Collection) Option {
	return func(o *opts) {
		o.coll = c
	}
}

func WithLocker(l Locker) Option {
	return func(o *opts) {
		o.locker = l
	}
}

func WithLogger(l Logger) Option {
	return func(o *opts) {
		o.logger = l
	}
}

var DefaultOptions = []Option{
	VersionSchemaName("schema_migrate"),
	VersionTableName("schema_version"),
	WithCollection(DefaulCollection),
	WithLocker(NewPostgresAdvisoryLocker(advisoryLockNumber)),
	WithLogger(NewSLogLogger(slog.Default())),
}

type opts struct {
	versionTableName  string
	versionSchemaName string
	coll              Collection
	locker            Locker
	logger            Logger
}

func applyOpts(os ...Option) *opts {
	opts := &opts{}
	for _, o := range append(DefaultOptions, os...) {
		o(opts)
	}

	return opts
}
