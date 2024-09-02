package clpgxmigrate

import (
	"context"
	"log/slog"
)

type Logger interface {
	LogNoVersionToApply(ctx context.Context, all, toApply []int64, currentVersion, targetVersion int64)
	LogVersionsToApply(ctx context.Context, all, toApply []int64, currentVersion, targetVersion int64)
	LogVersionApplyStart(ctx context.Context, version int64)
	LogVersionApplyDone(ctx context.Context, version int64)
}

func NewSLogLogger(logs *slog.Logger) Logger {
	return &slogLogger{logs}
}

type slogLogger struct {
	logs *slog.Logger
}

func (l *slogLogger) LogNoVersionToApply(ctx context.Context, all, _ []int64, currentVersion, targetVersion int64) {
	l.logs.InfoContext(ctx, "no versions to apply",
		slog.Int64("current_version", currentVersion),
		slog.Int64("target_version", targetVersion),
		slog.Any("available_versions", all),
	)
}

func (l *slogLogger) LogVersionsToApply(ctx context.Context, _, toApply []int64, currentVersion, targetVersion int64) {
	l.logs.InfoContext(ctx, "determined some versions to apply",
		slog.Int64("current_version", currentVersion),
		slog.Int64("target_version", targetVersion),
		slog.Any("version_to_apply", toApply),
	)
}

func (l *slogLogger) LogVersionApplyStart(ctx context.Context, version int64) {
	l.logs.DebugContext(ctx, "starting to apply migration", slog.Int64("version", version))
}

func (l *slogLogger) LogVersionApplyDone(ctx context.Context, version int64) {
	l.logs.DebugContext(ctx, "done applying migration", slog.Int64("version", version))
}
