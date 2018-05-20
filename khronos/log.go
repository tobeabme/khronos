package khronos

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
)

func init() {
	InitLogger("debug")
}

func InitLogger(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithError(err).Error("Error parsing log level, using: info")
		level = log.InfoLevel
	}
	log.SetLevel(level)
	initLogrus()
}

func initLogrus() {
	//set filename of log
	exeName := filepath.Base(os.Args[0])
	var extension = filepath.Ext(exeName)
	var logFileName = exeName[0 : len(exeName)-len(extension)]
	//set path of log
	logDir, err := GetCurrentPath()
	if err != nil {
		log.Error(err)
	}
	logPath := logDir + "/logs/"
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
		&log.JSONFormatter{}, // log.SetFormatter(&log.TextFormatter{})
	))

	// gin.DefaultWriter = log.Writer()

}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
