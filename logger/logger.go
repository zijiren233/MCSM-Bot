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
	FileOBJ    *os.File
	ErrFileOBJ *os.File
	Message    chan *Logmsg
}

type Logmsg struct {
	Levle   uint
	Message string
	Now     string
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
	logger.FileInit()
	logger.Message = make(chan *Logmsg, 500)
	go logger.backWriteLog()
	return &logger
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (l *Logger) FileInit() {
	if !Exists("./logs") {
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
	l.FileOBJ = f
	l.ErrFileOBJ = ef
}

func (l *Logger) BackupLog() {
	file, _ := l.FileOBJ.Stat()
	// 2M
	if file.Size() >= 2097152 {
		l.FileOBJ.Close()
		os.Rename(`./logs/log.log`, fmt.Sprint(`./logs/`, time.Now().Format("2006_01_02_15_04_05_bak_log.log")))
		f, _ := os.OpenFile("./logs/log.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		l.FileOBJ = f
	}
}

func (l *Logger) BackupErrLog() {
	file, _ := l.ErrFileOBJ.Stat()
	if file.Size() >= 2097152 {
		l.ErrFileOBJ.Close()
		os.Rename(`./logs/errlog.log`, fmt.Sprint(`./logs/`, time.Now().Format("2006_01_02_15_04_05_bak_errlog.log")))
		f, _ := os.OpenFile("./logs/errlog.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		l.ErrFileOBJ = f
	}
}

func (l *Logger) log(levle uint, format string, a ...interface{}) {
	if l.Levle <= levle {
		select {
		case l.Message <- &Logmsg{
			Levle:   levle,
			Message: fmt.Sprintf(format, a...),
			Now:     time.Now().Format("[2006-01-02 15:04:05] "),
		}:
		default:
		}
	}
}

func (l *Logger) backWriteLog() {
	var msgtmp *Logmsg
	for {
		msgtmp = <-l.Message
		l.BackupLog()
		fmt.Fprintln(l.FileOBJ, msgtmp.Now, fmt.Sprint("[", IntToLevle(msgtmp.Levle), "] "), msgtmp.Message)
		if msgtmp.Levle >= Error {
			l.BackupErrLog()
			fmt.Fprintln(l.ErrFileOBJ, msgtmp.Now, fmt.Sprint("[", IntToLevle(msgtmp.Levle), "] "), msgtmp.Message)
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
