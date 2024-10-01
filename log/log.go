package log

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"appserver/configures"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var infoLogger *logrus.Logger
var errorLogger *logrus.Logger

func InitLogs() {
	initErrorLogger()
	initInfoLogger()
}

func initInfoLogger() {
	infoLogger = logrus.New()
	_, err := rotatelogs.New(
		fmt.Sprintf(`%s/%s.%%Y%%m%%d.log`, configures.Config.Log.LogPath, configures.Config.Log.LogName),
		rotatelogs.WithLinkName(fmt.Sprintf(`%s/%s.log`, configures.Config.Log.LogPath, configures.Config.Log.LogName)),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithRotationSize(512*1024*1024),
	)
	if err != nil {
		log.Printf("init log error: %s", err)
		return
	}

	infoLogger.SetOutput(os.Stdout)
	infoLogger.SetReportCaller(true)

	infoLogger.SetFormatter(&LogFormatter{})
	infoLogger.SetLevel(logrus.DebugLevel)
}

type LogFormatter struct {
}

func (m *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("060102150405.000")
	newLog := fmt.Sprintf("%s\t%s\n", timestamp, entry.Message)
	b.WriteString(newLog)
	return b.Bytes(), nil
}

func initErrorLogger() {
	errorLogger = logrus.New()
	//writer
	_, err := rotatelogs.New(
		fmt.Sprintf(`%s/%s.%%Y%%m%%d.log`, configures.Config.Log.LogPath, configures.Config.Log.LogName+"_err"),
		rotatelogs.WithLinkName(fmt.Sprintf(`%s/%s.log`, configures.Config.Log.LogPath, configures.Config.Log.LogName+"_err")),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithRotationSize(512*1024*1024),
	)
	if err != nil {
		log.Printf("init log error: %s", err)
		return
	}

	//errorLogger.SetOutput(writer)
	errorLogger.SetOutput(os.Stdout)
	errorLogger.SetReportCaller(true)
	errorLogger.SetFormatter(&LogFormatter{})
	errorLogger.SetLevel(logrus.WarnLevel)
}

func Panic(f interface{}, v ...interface{}) {
	errorLogger.Panic(f, v)
}

func Fata(f interface{}, v ...interface{}) {
	errorLogger.Fatal(f, v)
}

func Error(f interface{}, v ...interface{}) {
	errorLogger.Error(f, v)
}

func Errorf(format string, v ...interface{}) {
	errorLogger.Errorf(format, v...)
}

func Warn(f interface{}, v ...interface{}) {
	errorLogger.Warn(f, v)
}

func Warnf(format string, v ...interface{}) {
	errorLogger.Warnf(format, v...)
}

func Info(v ...interface{}) {
	pl := len(v)
	if pl > 0 {
		arr := make([]string, pl)
		for i := 0; i < pl; i++ {
			arr[i] = "%v"
		}
		format := strings.Join(arr, "\t")
		infoLogger.Info(fmt.Sprintf(format, v...))
	}
}

func Infof(format string, v ...interface{}) {
	infoLogger.Info(fmt.Sprintf(format, v...))
}

func Debugf(format string, v ...interface{}) {
	infoLogger.Debug(fmt.Sprintf(format, v...))
}

func Tracef(format string, v ...interface{}) {
	infoLogger.Trace(fmt.Sprintf(format, v...))
}

type LogEntity struct {
	fields map[string]interface{}
}

func WithContext(ctx context.Context) *LogEntity {
	log := &LogEntity{
		fields: map[string]interface{}{},
	}
	//handle ctx
	return log
}

func (log *LogEntity) WithField(key string, value interface{}) *LogEntity {
	log.fields[key] = value
	return log
}

func (log *LogEntity) Errorf(format string, v ...interface{}) {
	arr := []interface{}{}
	initFormat := ""
	for k, v := range log.fields {
		initFormat = initFormat + k + ":" + "%v\t"
		arr = append(arr, v)
	}
	arr = append(arr, v...)
	Errorf(initFormat+format, arr...)
}

func (log *LogEntity) Error(errMsg string) {
	log.Errorf(errMsg)
}

func (log *LogEntity) Warnf(format string, v ...interface{}) {
	arr := []interface{}{}
	initFormat := ""
	for k, v := range log.fields {
		initFormat = initFormat + k + ":" + "%v\t"
		arr = append(arr, v)
	}
	arr = append(arr, v...)
	Warnf(initFormat+format, arr...)
}

func (log *LogEntity) Warn(warnMsg string) {
	log.Warnf(warnMsg)
}

func (log *LogEntity) Infof(format string, v ...interface{}) {
	arr := []interface{}{}
	initFormat := ""
	for k, v := range log.fields {
		initFormat = initFormat + k + ":" + "%v\t"
		arr = append(arr, v)
	}
	arr = append(arr, v...)
	Infof(initFormat+format, arr...)
}

func (log *LogEntity) Info(infoMsg string) {
	log.Infof(infoMsg)
}
