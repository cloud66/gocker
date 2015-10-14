package cxlogger

// see http://godoc.org/github.com/inconshreveable/log15 for more info
import log "gopkg.in/inconshreveable/log15.v2"

const tabWidth = 2

var contextIndentations = make(map[string]int)

// func main() {
// 	Initialize("STDOUT", "debug")
// 	Debug("test!")
// }

const (
	LvlCrit  = log.LvlCrit
	LvlError = log.LvlError
	LvlWarn  = log.LvlWarn
	LvlInfo  = log.LvlInfo
	LvlDebug = log.LvlDebug
)

var Log *Logger

func Initialize(logOut string, lvl interface{}) error {
	if Log == nil {
		Log = &Logger{
			Logger: log.New(),
		}
	}
	return Log.InitializeWithContext("main", logOut, lvl)
}

func Debug(v ...interface{}) { Log.Debug(v...) }
func Info(v ...interface{})  { Log.Info(v...) }
func Warn(v ...interface{})  { Log.Warn(v...) }
func Error(v ...interface{}) { Log.Error(v...) }
func Crit(v ...interface{})  { Log.Crit(v...) }

func Debugf(format string, v ...interface{}) { Log.Debugf(format, v...) }
func Infof(format string, v ...interface{})  { Log.Infof(format, v...) }
func Warnf(format string, v ...interface{})  { Log.Warnf(format, v...) }
func Errorf(format string, v ...interface{}) { Log.Errorf(format, v...) }
func Critf(format string, v ...interface{})  { Log.Critf(format, v...) }

func DebugIndent(indentation int, v ...interface{}) { Log.DebugIndent(indentation, v) }
func InfoIndent(indentation int, v ...interface{})  { Log.InfoIndent(indentation, v) }
func WarnIndent(indentation int, v ...interface{})  { Log.WarnIndent(indentation, v) }
func ErrorIndent(indentation int, v ...interface{}) { Log.ErrorIndent(indentation, v) }
func CritIndent(indentation int, v ...interface{})  { Log.CritIndent(indentation, v) }

func IncreaseIndentation() { Log.IncreaseIndentation() }
func DecreaseIndentation() { Log.DecreaseIndentation() }
