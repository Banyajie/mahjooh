package logger

import (
	"chess_alg_jx/config"
	"github.com/op/go-logging"
	"os"
)

var logLevel logging.Level

var logg = logging.MustGetLogger("chess")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} %{level:.4s} > %{color:reset} %{message}`,
)

func InitLog() {
	switch config.Config.LogLevel {
	case "info":
		logLevel = logging.INFO
	case "debug":
		logLevel = logging.DEBUG
	case "notice":
		logLevel = logging.NOTICE
	case "warning":
		logLevel = logging.WARNING
	case "error":
		logLevel = logging.ERROR
	}

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backendLeveled, backendFormatter)
}

func Debug(a ...interface{}) {
	if logLevel >= logging.DEBUG {
		logg.Debug(a...)
	}
}

func Debugf(format string, a ...interface{}) {
	if logLevel >= logging.DEBUG {
		logg.Debugf(format, a...)
	}
}

func Info(a ...interface{}) {
	if logLevel >= logging.INFO {
		logg.Info(a...)
	}
}

func Infof(format string, a ...interface{}) {
	if logLevel >= logging.INFO {
		logg.Infof(format, a...)
	}
}

func Notice(a ...interface{}) {
	if logLevel >= logging.NOTICE {
		logg.Notice(a...)
	}
}

func Noticef(format string, a ...interface{}) {
	if logLevel >= logging.NOTICE {
		logg.Noticef(format, a...)
	}
}

func Warning(a ...interface{}) {
	if logLevel >= logging.WARNING {
		logg.Warning(a...)
	}
}

func Warningf(format string, a ...interface{}) {
	if logLevel >= logging.WARNING {
		logg.Warningf(format, a...)
	}
}

func Error(a ...interface{}) {
	if logLevel >= logging.ERROR {
		logg.Error(a...)
	}
}

func Errorf(format string, a ...interface{}) {
	if logLevel >= logging.ERROR {
		logg.Errorf(format, a...)
	}
}

func Fatal(a ...interface{}) {
	logg.Fatal(a...)
}

func Fatalf(format string, a ...interface{}) {
	logg.Fatalf(format, a...)
}
