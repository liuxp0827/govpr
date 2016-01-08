package log

func init() {
	logger = NewLogger(10000)
	logger.SetLogFuncCallWithDepth(true, 3)
	logger.SetLogger("console", "")
	logger.SetLevel(LevelDebug)
}

// logger references the used application logger.
var logger *Logger

// SetLogLevel sets the global log level used by the simple
// logger.
func SetLevel(l int) {
	logger.SetLevel(l)
}

func GetDefaultLogger()*Logger{
	return logger
}

func SetLogFuncCall(b bool) {
	logger.SetLogFuncCall(b)
}

func SetLogFile(logFile string, level int, isRotateDaily, drawColor bool, rotateMaxDays int) {
	logger.SetLogFile(logFile, level, isRotateDaily, drawColor, rotateMaxDays)
}

// Error logs a message at error level.
func Error(v ...interface{}) {
	logger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Warning logs a message at warning level.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

func Info(v ...interface{}) {
	logger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Debug logs a message at debug level.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

func Trace(v ...interface{}) {
	logger.Trace(v...)
}

func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}
