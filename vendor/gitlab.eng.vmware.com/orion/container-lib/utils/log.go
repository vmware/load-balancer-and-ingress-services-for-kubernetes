package utils

import (
	"io/ioutil"
	"log"
	"os"
)

type AviLogger struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

var AviLog AviLogger

func init() {
	// TODO (sudswas): evaluate if moving to a Regular function is better than package init)
	// Change from ioutil.Discard for log to appear
	AviLog.Trace = log.New(ioutil.Discard,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	AviLog.Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	AviLog.Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	AviLog.Error = log.New(os.Stdout,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
