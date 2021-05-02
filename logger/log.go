package logger

import (
	"fmt"
	"github.com/brahma-adshonor/gohook"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
	_ "unsafe"
)

// InitLog logName 日志文件名
// logLevel日志级别
// maxRemainCnt 设置文件清理前最多保存的个数
func InitLog(logName, logLevel *string, maxRemainNum uint) {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&LogFormatter{})
	hook := newLfsHook(*logName, *logLevel, maxRemainNum)
	logrus.AddHook(hook)
}

type LogFormatter struct {
}

func (s *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	var file string
	var len int
	if entry.Caller != nil {
		file = filepath.Base(entry.Caller.File)
		len = entry.Caller.Line
	}
	msg := fmt.Sprintf("%s [%s:%d][GOID:%d][%s] %s\n", timestamp, file, len, 1, strings.ToUpper(entry.Level.String()), entry.Message)
	return []byte(msg), nil
}

func newLfsHook(logName, logLevel string, maxRemainNum uint) logrus.Hook {
	if logName == "" {
		logName = "hdfs-mount.log"
		logrus.Warnf("logName is not set,default logName is hdfs-mount.log")
	}
	if logLevel == "" {
		logLevel = "info"
		logrus.Warnf("logger level is not set,default logger level is info")
	}
	if maxRemainNum == 0 {
		maxRemainNum = 365
		logrus.Warnf("logger maxRemainNum is not set,default logger maxRemainNum is 365")
	}
	if maxRemainNum < 0 {
		maxRemainNum = 365
		logrus.Warnf("logger maxRemainNum is invalid,maxRemainNum less than zero,maxRemainNum:%d,"+
			"will use default logger maxRemainNum is 365", maxRemainNum)
	}
	writer, err := rotatelogs.New(
		logName+".%Y%m%d",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		rotatelogs.WithLinkName(logName),
		// WithRotationTime设置日志分割的时间，这里设置为一小时分割一次
		rotatelogs.WithRotationTime(time.Hour*24),
		// WithMaxAge和WithRotationCount二者只能设置一个，
		// WithMaxAge 设置文件清理前的最长保存时间，
		// WithRotationCount 设置文件清理前最多保存的个数。
		// rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationCount(maxRemainNum),
	)

	if err != nil {
		logrus.Fatalf("config local file system for logger error: %v", err)
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatalf("config logger level error: %v,support level [panic fatal error warn info debug trace]", err)
	}

	logrus.SetLevel(level)

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &LogFormatter{})

	return lfsHook
}

//go:linkname Printf log.Printf
func Printf(format string, v ...interface{})

func ReplacePrintf(format string, v ...interface{}) {
	logrus.Debugf(format, v)
}

//go:linkname Println log.Println
func Println(v ...interface{})
func ReplacePrintln(v ...interface{}) {
	logrus.Debug(v)
}
func init() {
	err := gohook.HookByIndirectJmp(Printf, ReplacePrintf, nil)
	if err != nil {
		logrus.Error("replace log to logrus fail")
	}
	err = gohook.HookByIndirectJmp(Println, ReplacePrintln, nil)
	if err != nil {
		logrus.Error("replace log to logrus fail")
	}
}
