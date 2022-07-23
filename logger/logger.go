package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/zijiren233/MCSM-Bot/utils"
)

var log Logger

const (
	Debug uint = iota
	Info
	Warning
	Error
	Fatal
	None
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type Logger struct {
	levle        uint
	disableprint bool
	fileOBJ      *os.File
	errFileOBJ   *os.File
	message      chan *logmsg
}

type logmsg struct {
	levle    uint
	message  string
	now      string
	funcName string
	filename string
	line     int
}

func levleColor(levle uint) string {
	switch levle {
	case Debug:
		return blue
	case Info:
		return green
	case Warning:
		return yellow
	case Error:
		return red
	case Fatal:
		return magenta
	default:
		return red
	}
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
	case "fatal":
		return Fatal
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
	case Fatal:
		return "FATAL"
	case None:
		return "NONE"
	default:
		return "DEBUG"
	}
}

func init() {
	log = Logger{levle: Info, disableprint: false}
	log.fileInit()
	log.message = make(chan *logmsg, 500)
	go log.backWriteLog()
}

func GetLog() *Logger {
	return &log
}

func DisableLogPrint() {
	log.disableprint = true
}

func EnableLogPrint() {
	log.disableprint = false
}

func (l *Logger) fileInit() {
	if !utils.FileExists("./logs") {
		os.Mkdir("./logs", os.ModePerm)
	}
	f, err := os.OpenFile("./logs/log.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("打开日志错误!")
		panic(err)
	}
	ef, err2 := os.OpenFile("./logs/errlog.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("打开Err日志错误!")
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
	filename, funcName, line := getInfo()
	if l.levle <= levle {
		select {
		case l.message <- &logmsg{
			levle:    levle,
			message:  fmt.Sprintf(format, a...),
			now:      time.Now().Format("[2006-01-02 15:04:05] "),
			funcName: funcName,
			filename: filename,
			line:     line,
		}:
		default:
		}
	}
}

func (l *Logger) SetLogLevle(levle uint) {
	l.levle = levle
}

func (l *Logger) backWriteLog() {
	var msgtmp *logmsg
	stdout := colorable.NewColorableStdout
	for {
		msgtmp = <-l.message
		l.backupLog()
		fmt.Fprintf(l.fileOBJ, "%s[%s] [%s|%s|%d] %s\n", msgtmp.now, IntToLevle(msgtmp.levle), msgtmp.filename, msgtmp.funcName, msgtmp.line, msgtmp.message)
		if !l.disableprint {
			fmt.Fprintf(stdout(), "%s|%s %s %s| %s\n", msgtmp.now, levleColor(msgtmp.levle), IntToLevle(msgtmp.levle), reset, msgtmp.message)
		}
		if msgtmp.levle >= Error {
			l.backupErrLog()
			fmt.Fprintf(l.errFileOBJ, "%s[%s] [%s|%s|%d] %s\n", msgtmp.now, "ERROR", msgtmp.filename, msgtmp.funcName, msgtmp.line, msgtmp.message)
		}
	}
}

func getInfo() (string, string, int) {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "", "", 0
	}
	return path.Base(file), path.Base(runtime.FuncForPC(pc).Name()), line
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

func (l *Logger) Fatal(format string, a ...interface{}) {
	l.log(Fatal, format, a...)
}
