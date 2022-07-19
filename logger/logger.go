package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var Log *Logger

const (
	Debug uint = iota
	Info
	Warning
	Error
	None
)

type Logger struct {
	Levle      uint
	fileOBJ    *os.File
	errFileOBJ *os.File
	message    chan *logmsg
}

type logmsg struct {
	levle   uint
	message string
	now     string
}

func LevleToInt(s string) uint {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warning":
		return Warning
	case "error":
		return Error
	case "none":
		return None
	default:
		return Debug
	}
}

func IntToLevle(i uint) string {
	switch i {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case None:
		return "NONE"
	default:
		return "DEBUG"
	}
}

func Newlog(levle uint) *Logger {
	logger := Logger{Levle: levle}
	logger.fileInit()
	logger.message = make(chan *logmsg, 500)
	go logger.backWriteLog()
	return &logger
}

func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (l *Logger) fileInit() {
	if !exists("./logs") {
		os.Mkdir("./logs", os.ModePerm)
	}
	f, err := os.OpenFile("./logs/log.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("打开日志错误！")
		panic(err)
	}
	ef, err2 := os.OpenFile("./logs/errlog.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("打开Err日志错误！")
		panic(err2)
	}
	l.fileOBJ = f
	l.errFileOBJ = ef
}

func (l *Logger) backupLog() {
	file, _ := l.fileOBJ.Stat()
	// 2M
	if file.Size() >= 2097152 {
		l.fileOBJ.Close()
		os.Rename(`./logs/log.log`, fmt.Sprint(`./logs/`, time.Now().Format("2006_01_02_15_04_05_bak_log.log")))
		f, _ := os.OpenFile("./logs/log.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		l.fileOBJ = f
	}
}

func (l *Logger) backupErrLog() {
	file, _ := l.errFileOBJ.Stat()
	if file.Size() >= 2097152 {
		l.errFileOBJ.Close()
		os.Rename(`./logs/errlog.log`, fmt.Sprint(`./logs/`, time.Now().Format("2006_01_02_15_04_05_bak_errlog.log")))
		f, _ := os.OpenFile("./logs/errlog.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		l.errFileOBJ = f
	}
}

func (l *Logger) log(levle uint, format string, a ...interface{}) {
	if l.Levle <= levle {
		select {
		case l.message <- &logmsg{
			levle:   levle,
			message: fmt.Sprintf(format, a...),
			now:     time.Now().Format("[2006-01-02 15:04:05] "),
		}:
		default:
		}
	}
}

func (l *Logger) backWriteLog() {
	var msgtmp *logmsg
	for {
		msgtmp = <-l.message
		l.backupLog()
		fmt.Fprintln(l.fileOBJ, msgtmp.now, fmt.Sprint("[", IntToLevle(msgtmp.levle), "] "), msgtmp.message)
		if msgtmp.levle >= Error {
			l.backupErrLog()
			fmt.Fprintln(l.errFileOBJ, msgtmp.now, fmt.Sprint("[", IntToLevle(msgtmp.levle), "] "), msgtmp.message)
		}
	}
}

func (l *Logger) Debug(format string, a ...interface{}) {
	l.log(Debug, format, a...)
}

func (l *Logger) Info(format string, a ...interface{}) {
	l.log(Info, format, a...)
}

func (l *Logger) Warring(format string, a ...interface{}) {
	l.log(Warning, format, a...)
}

func (l *Logger) Error(format string, a ...interface{}) {
	l.log(Error, format, a...)
}
