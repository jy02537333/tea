package logx

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

// Init 初始化zap日志
func Init() error {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	l, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = l
	// 将其设置为全局logger，兼容 zap.L()
	zap.ReplaceGlobals(l)
	return nil
}

func Get() *zap.Logger {
	if logger == nil {
		_ = Init()
	}
	return logger
}

// Sync 在进程结束时调用
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
