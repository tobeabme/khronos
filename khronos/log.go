package khronos

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// var log = logrus.New()

func InitLogger(logLevel string, logPath string) {
	initLogrus(logLevel, logPath)
}

func initLogrus(logLevel string, logPath string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithError(err).Error("Error parsing log level, using: info")
		level = log.InfoLevel
	}
	log.SetLevel(level)

	if logPath == "stdout" {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stdout)
		return
	}

	//set filename of log
	exeName := filepath.Base(os.Args[0])
	var extension = filepath.Ext(exeName)
	var logFileName = exeName[0 : len(exeName)-len(extension)]

	//set path of log
	logFilePath := path.Join(logPath, logFileName)

	//create the filepath if dir is not exist
	dir := path.Dir(logFilePath)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		errInfo := fmt.Sprintf("create log dir error=%v", err)
		fmt.Println(errInfo)
		panic(errInfo)
	}

	writer, err := rotatelogs.New(
		logFilePath+".%Y%m%d",
		rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", err.Error())
	}
	log.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			log.DebugLevel: writer,
			log.InfoLevel:  writer,
			log.WarnLevel:  writer,
			log.ErrorLevel: writer,
			log.FatalLevel: writer,
			log.PanicLevel: writer,
		},
		&log.TextFormatter{}, // log.SetFormatter(&log.TextFormatter{})
		//&log.JSONFormatter{},
	))

	// gin.DefaultWriter = log.Writer()

}
