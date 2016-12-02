package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

// RFC5424 log message levels.
const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

type loggerType func() LoggerInterface

// LoggerInterface defines the behavior of a log provider.
type LoggerInterface interface {
	Init(config string) error
	WriteMsg(msg string, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]loggerType)

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log loggerType) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// Logger is default logger in beego application.
// it can contain several providers and log message into all providers.
type Logger struct {
	lock                sync.Mutex
	level               int
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	asynchronous        bool
	msg                 chan *logMsg
	outputs             map[string]LoggerInterface
}

type logMsg struct {
	level int
	msg   string
}

// NewLogger returns a new Logger.
// channellen means the number of messages in chan.
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channellen int64) *Logger {
	bl := new(Logger)
	bl.level = LevelDebug
	bl.loggerFuncCallDepth = 3
	bl.msg = make(chan *logMsg, channellen)
	bl.outputs = make(map[string]LoggerInterface)
	return bl
}

func NewConsoleLogger(channellen int64) *Logger {
	bl := new(Logger)
	bl.level = LevelDebug
	bl.loggerFuncCallDepth = 3
	bl.msg = make(chan *logMsg, channellen)
	bl.outputs = make(map[string]LoggerInterface)
	bl.SetLogger("console", "")
	return bl
}

func (bl *Logger) Async() *Logger {
	bl.asynchronous = true
	go bl.startLogger()
	return bl
}

// SetLogger provides a given logger adapter into Logger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *Logger) SetLogger(adaptername string, config string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if log, ok := adapters[adaptername]; ok {
		lg := log()
		err := lg.Init(config)
		bl.outputs[adaptername] = lg
		if err != nil {
			fmt.Println("logs.Logger.SetLogger: " + err.Error())
			return err
		}
	} else {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adaptername)
	}
	return nil
}

func (bl *Logger) SetLogFile(logFile string, level int, isRotateDaily, drawColor bool, rotateMaxDays int) error {
	return bl.SetLogger("file", fmt.Sprintf(`{"filename":"%s","level":%d,"daily":%v,"maxdays":%d,"drawcolor":%v}`, logFile, level, isRotateDaily, rotateMaxDays, drawColor))
}

// remove a logger adapter in Logger.
func (bl *Logger) DelLogger(adaptername string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if lg, ok := bl.outputs[adaptername]; ok {
		lg.Destroy()
		delete(bl.outputs, adaptername)
		return nil
	} else {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adaptername)
	}
}

func (bl *Logger) writerMsg(loglevel int, msg string) error {
	lm := new(logMsg)
	lm.level = loglevel
	if bl.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(bl.loggerFuncCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		lm.msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
	} else {
		lm.msg = msg
	}
	if bl.asynchronous {
		bl.msg <- lm
	} else {
		for name, l := range bl.outputs {
			err := l.WriteMsg(lm.msg, lm.level)
			if err != nil {
				fmt.Println("unable to WriteMsg to adapter:", name, err)
				return err
			}
		}
	}
	return nil
}

// Set log message level.
//
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not even be sent the message.
func (bl *Logger) SetLevel(l int) {
	bl.level = l
}

// set log funcCallDepth
func (bl *Logger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// get log funcCallDepth for wrapper
func (bl *Logger) GetLogFuncCallDepth() int {
	return bl.loggerFuncCallDepth
}

// enable log funcCallDepth
func (bl *Logger) EnableFuncCallDepth(b bool) {
	bl.enableFuncCallDepth = b
}

func (bl *Logger) SetLogFuncCallWithDepth(b bool, depth int) {
	bl.EnableFuncCallDepth(b)
	bl.SetLogFuncCallDepth(depth)
}

func (bl *Logger) SetLogFuncCall(b bool) {
	bl.EnableFuncCallDepth(b)
	bl.SetLogFuncCallDepth(3)
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *Logger) startLogger() {
	for {
		select {
		case bm := <-bl.msg:
			for _, l := range bl.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg:", err)
				}
			}
		}
	}
}

// Log ERROR level message.
func (bl *Logger) Errorf(format string, v ...interface{}) {
	if LevelError > bl.level {
		return
	}
	msg := fmt.Sprintf("[E] "+format, v...)
	bl.writerMsg(LevelError, msg)
}

func (bl *Logger) Error(v ...interface{}) {
	if LevelError > bl.level {
		return
	}
	msg := "[E] " + fmt.Sprintf(generateFmtStr(len(v)), v...)
	bl.writerMsg(LevelError, msg)
}

// Log FATAL level message.
func (bl *Logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf("[F] "+format, v...)
	bl.writerMsg(LevelError, msg)
	os.Exit(1)
}

func (bl *Logger) Fatal(v ...interface{}) {
	msg := "[F] " + fmt.Sprintf(generateFmtStr(len(v)), v...)
	bl.writerMsg(LevelError, msg)
	os.Exit(1)
}

// Log WARNING level message.
func (bl *Logger) Warnf(format string, v ...interface{}) {
	if LevelWarn > bl.level {
		return
	}
	msg := fmt.Sprintf("[W] "+format, v...)
	bl.writerMsg(LevelWarn, msg)
}

func (bl *Logger) Warn(v ...interface{}) {
	if LevelWarn > bl.level {
		return
	}
	msg := "[W] " + fmt.Sprintf(generateFmtStr(len(v)), v...)
	bl.writerMsg(LevelWarn, msg)
}

// Log INFORMATIONAL level message.
func (bl *Logger) Infof(format string, v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	msg := fmt.Sprintf("[I] "+format, v...)
	bl.writerMsg(LevelInfo, msg)
}

func (bl *Logger) Info(v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	msg := "[I] " + fmt.Sprintf(generateFmtStr(len(v)), v...)
	bl.writerMsg(LevelInfo, msg)
}

// Log DEBUG level message.
func (bl *Logger) Debugf(format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	msg := fmt.Sprintf("[D] "+format, v...)
	bl.writerMsg(LevelDebug, msg)
}

func (bl *Logger) Debug(v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	msg := "[D] " + fmt.Sprintf(generateFmtStr(len(v)), v...)
	bl.writerMsg(LevelDebug, msg)
}

// Log TRACE level message.
// compatibility alias for Debug()
func (bl *Logger) Tracef(format string, v ...interface{}) {
	if LevelTrace > bl.level {
		return
	}
	msg := fmt.Sprintf("[T] "+format, v...)
	bl.writerMsg(LevelTrace, msg)
}

func (bl *Logger) Trace(v ...interface{}) {
	if LevelTrace > bl.level {
		return
	}
	msg := "[T] " + fmt.Sprintf(generateFmtStr(len(v)), v...)
	bl.writerMsg(LevelTrace, msg)
}

// flush all chan data.
func (bl *Logger) Flush() {
	for _, l := range bl.outputs {
		l.Flush()
	}
}

// close logger, flush all chan data and destroy all adapters in Logger.
func (bl *Logger) Close() {
	for {
		if len(bl.msg) > 0 {
			bm := <-bl.msg
			for _, l := range bl.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg (while closing logger):", err)
				}
			}
			continue
		}
		break
	}
	for _, l := range bl.outputs {
		l.Flush()
		l.Destroy()
	}
}

func generateFmtStr(n int) string {
	return strings.Repeat("%v ", n)
}
