package core

type Logger interface {

	//basic log method
	Log(prefix string, format string, v ...any)

	//log with different level.
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
	Panic(format string, v ...any)
}
