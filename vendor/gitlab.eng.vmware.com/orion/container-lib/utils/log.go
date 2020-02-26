package utils

import (
	"io/ioutil"
	"log"
	"os"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type AviLogger struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

var AviLog AviLogger

const (
	InfoColor  = "\033[1;32mINFO: \033[0m"
	WarnColor  = "\033[1;33mWARNING: \033[0m"
	ErrColor   = "\033[1;31mERROR: \033[0m"
	TraceColor = "\033[0;36mTRACE: \033[0m"
)

func init() {
	// Change from ioutil.Discard for log to appear
	// User provides the log file and lumberjack should create compressed log files with the naming format:<file_name>-2020-11-01sT18-30-00.000.log.gz post rotation.
	var file *os.File
	var logpath string
	var err error
	usePVC := os.Getenv("USE_PVC")
	if usePVC == "true" {
		var logpath = os.Getenv("LOG_FILE_NAME")
		if logpath == "" {
			logpath = DEFAULT_AVI_LOG
		}
		file, err = os.OpenFile(logpath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
	} else {
		file = os.Stdout
	}

	AviLog.Trace = log.New(ioutil.Discard,
		TraceColor,
		log.Ldate|log.Ltime|log.Lshortfile)

	AviLog.Info = log.New(file,
		InfoColor,
		log.Ldate|log.Ltime|log.Lshortfile)
	if usePVC == "true" {
		AviLog.Info.SetOutput(&lumberjack.Logger{
			Filename:   logpath,
			MaxSize:    1,  // megabytes after which new file is created
			MaxBackups: 3,  // number of backups
			MaxAge:     28, //days
			Compress:   true,
		})
	}

	AviLog.Warning = log.New(file,
		WarnColor,
		log.Ldate|log.Ltime|log.Lshortfile)

	AviLog.Error = log.New(file,
		ErrColor,
		log.Ldate|log.Ltime|log.Lshortfile)
}
