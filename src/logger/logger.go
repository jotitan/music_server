package logger

import (
	"log"
	"os"
	"runtime"
	"strings"
	"fmt"
	"sync"
)

/* Manage log */
/* If file is defined, write inside. In any case, write in console */

// Logger manage log production
type Logger struct{
	info * log.Logger
	error * log.Logger
	writeConsole bool
}

func getInfo() string {
	_, file, line, _ := runtime.Caller(3)
	return fmt.Sprintf("%s:%d:", file[strings.LastIndex(file, "/")+1:], line)
}

// Info write info into log
func (l Logger) Info(message ...interface{}) {
	l.print(l.info, message...)
}

// Fatal write fatal info into log
func (l Logger) Fatal(message ...interface{}) {
	l.print(l.error, message...)
	os.Exit(1)
}

// Erro write error message into log
func (l Logger) Error(message ...interface{}) {
	l.print(l.error, message...)
}

// pring write message into logger. If console is enabled, write into too
func (l Logger) print(loggerElement * log.Logger, message ...interface{}) {
	data := append([]interface{}{ getInfo()}, message...)
	loggerElement.Println(data...)
	if l.writeConsole {
		log.Println(data...)
	}
}

// InitLogger init the logger which will write messages into filename file and / or console
func InitLogger(filename string, console bool) *Logger {
	out := os.Stdout
	errOut := os.Stdout
	logger = &Logger{writeConsole:false}
	if filename != "" {
		if file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm) ; err == nil {
			out = file
			errOut = file
			logger.writeConsole = console
		}else{
			fmt.Println("Error when creating logger",err)
		}
	}
	logger.info = log.New(out, "INFO ", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.error = log.New(errOut, "ERROR ", log.Ldate|log.Ltime)
	return logger
}

// singleton of logger
var logger *Logger
var lock = sync.Mutex{}

// GetLogger return the logger or create it if not exist
// TODO singleton, synchronize it
func GetLogger() *Logger {
	if logger == nil {
		lock.Lock()
		if logger == nil {
			InitLogger("", false)
		}
		lock.Unlock()
	}
	return logger
}
