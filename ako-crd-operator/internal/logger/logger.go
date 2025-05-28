package logger

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
	"time"
)

type Logger struct {
	InnerLogger *zap.SugaredLogger
}

func logLevel() zapcore.Level {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

const defaultFileSuffix = "avi.log"

// log file sample name /log/ako-12345.avi.log
func getFileName() string {
	input := os.Getenv("LOG_FILE_NAME")
	if input == "" {
		input = defaultFileSuffix
	}
	fileName := fmt.Sprintf("%s%s%s.%d", getFilePath(), getPodName(), input, time.Now().Unix())
	return fileName
}

func getFilePath() string {
	return strings.TrimLeft(os.Getenv("LOG_FILE_PATH")+"/", "/")
}

func getPodName() string {
	return strings.TrimLeft(os.Getenv("POD_NAME")+".", ".")
}

func NewLogger() *Logger {
	atom := zap.NewAtomicLevel()
	// default level set to Info
	atom.SetLevel(logLevel())

	var file *os.File
	var logpath string
	var err error

	usePVC := os.Getenv("USE_PVC")

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // colored capital case LEVEL
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder        // format 2020-05-08T03:26:08.943+0530
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder      // caller format package_name/filename.go

	if usePVC != "true" {
		logger := zap.New(zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atom,
		))

		logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
		sugar := logger.Sugar()
		return &Logger{sugar}

	}

	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	logpath = getFileName()
	file, err = os.OpenFile(logpath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	file.Close()

	level := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= atom.Level()
	})
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logpath,
		MaxSize:    500, // megabytes after which new file is created
		MaxBackups: 5,   // number of backups
		MaxAge:     28,  // days
		Compress:   true,
	})
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg),
		w,
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := logger.Sugar()
	defer sugar.Sync()
	return &Logger{sugar}
}

// Init implements logr.LogSink.
func (l Logger) Init(info logr.RuntimeInfo) {
	// Not used
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.InnerLogger.Infof(template, args...)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.InnerLogger.Info(msg)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.InnerLogger.Warnf(template, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.InnerLogger.Warn(args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.InnerLogger.Errorf(template, args...)
}

func (l *Logger) Error(err error, msg string, args ...interface{}) {
	l.InnerLogger.Error(msg)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.InnerLogger.Debugf(template, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.InnerLogger.Debug(args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.InnerLogger.Fatal(args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.InnerLogger.Fatalf(template, args...)
}

func (l *Logger) WithValues(keysAndValues ...interface{}) *Logger {
	return &Logger{l.InnerLogger.With(keysAndValues...)}
}

func (l *Logger) WithName(name string) *Logger {
	return &Logger{l.InnerLogger.Named(name)}
}

type loggerkey string

var key loggerkey = "logger"

func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(key).(*Logger)
	if !ok {
		return NewLogger()
	}
	return logger
}

func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}
