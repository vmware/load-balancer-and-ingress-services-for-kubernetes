/*
 * Copyright 2019-2020 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
)

var LogLevelMap = map[string]zapcore.Level{
	"DEBUG": DebugLevel,
	"INFO":  InfoLevel,
	"WARN":  WarnLevel,
	"ERROR": ErrorLevel,
}

type AviLogger struct {
	sugar  *zap.SugaredLogger
	logger *zap.Logger // sugar is obtained from this logger
	// Sugaring a Logger is quite inexpensive, so it's reasonable for a single application to use both Loggers and SugaredLoggers, converting between them on the boundaries of performance-sensitive code.
	atom zap.AtomicLevel
}

// Init implements logr.LogSink.
func (AviLogger) Init(info logr.RuntimeInfo) {
	// Not used
}

func (aviLogger *AviLogger) Infof(template string, args ...interface{}) {
	aviLogger.sugar.Infof(template, args...)
}

func (aviLogger AviLogger) Info(level int, msg string, args ...interface{}) {
	aviLogger.sugar.Info(msg)
}

func (aviLogger *AviLogger) Warnf(template string, args ...interface{}) {
	aviLogger.sugar.Warnf(template, args...)
}

func (aviLogger *AviLogger) Warn(args ...interface{}) {
	aviLogger.sugar.Warn(args...)
}

func (aviLogger *AviLogger) Errorf(template string, args ...interface{}) {
	aviLogger.sugar.Errorf(template, args...)
}

func (aviLogger AviLogger) Error(err error, msg string, args ...interface{}) {
	aviLogger.sugar.Error(msg)
}

func (aviLogger *AviLogger) Debugf(template string, args ...interface{}) {
	aviLogger.sugar.Debugf(template, args...)
}

func (aviLogger *AviLogger) Debug(args ...interface{}) {
	aviLogger.sugar.Debug(args...)
}

func (aviLogger *AviLogger) Fatal(args ...interface{}) {
	aviLogger.sugar.Fatal(args...)
}

func (aviLogger *AviLogger) Fatalf(template string, args ...interface{}) {
	aviLogger.sugar.Fatalf(template, args...)
}

// SetLevel changes loglevel during runtime
func (aviLogger *AviLogger) SetLevel(l string) {
	aviLogger.atom.SetLevel(LogLevelMap[l])
}

func (aviLogger AviLogger) Enabled(level int) bool {
	return aviLogger.sugar != nil
}

func (aviLogger AviLogger) V(level int) logr.LogSink {
	return aviLogger
}

func (aviLogger AviLogger) WithValues(keysAndValues ...interface{}) logr.LogSink {
	// Not used
	return &aviLogger
}

func (aviLogger AviLogger) WithName(name string) logr.LogSink {
	_ = aviLogger.sugar.Named(name)
	return &aviLogger

}

// log file sample name /log/ako-12345.avi.log
func getFileName() string {
	input := os.Getenv("LOG_FILE_NAME")
	if input == "" {
		input = DEFAULT_FILE_SUFFIX
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

var AviLog AviLogger

func init() {
	atom := zap.NewAtomicLevel()
	// default level set to Info
	atom.SetLevel(InfoLevel)

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
		AviLog = AviLogger{sugar, logger, atom}
		return
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
	AviLog = AviLogger{sugar, logger, atom}
}
