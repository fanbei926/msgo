package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type LoggerLevel int

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelError:
		return "error"
	default:
		return ""
	}
}

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

type LoggerFormatter struct {
	Level   LoggerLevel
	IsColor bool
}

type Logger struct {
	Formatter LoggerFormatter
	Level     LoggerLevel
	Outs      []io.Writer
}

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	logger.Level = LevelDebug
	logger.Outs = append(logger.Outs, os.Stdout)
	return logger
}

func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}

func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}

func (l *Logger) Print(level LoggerLevel, msg any) {
	if l.Level > level {
		// current level is higher than input level, so do not print it
		return
	}
	l.Formatter.Level = level
	str := l.Formatter.format(msg)
	for _, out := range l.Outs {
		if out == os.Stdout {
			l.Formatter.IsColor = true
			str = l.Formatter.format(msg)
		}
		fmt.Fprintln(out, str)
	}
}

func (f *LoggerFormatter) format(msg any) string {
	now := time.Now()
	if f.IsColor {
		//要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [msgo] %s %s%v%s | level= %s %s %s | msg=%s %#v %s \n",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, f.Level.Level(), reset, msgColor, msg, reset,
		)
	}
	return fmt.Sprintf("[msgo] %v | level=%s | msg=%#v \n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(),
		msg,
	)
}

func (f *LoggerFormatter) LevelColor() interface{} {
	switch f.Level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (f *LoggerFormatter) MsgColor() interface{} {
	switch f.Level {
	case LevelError:
		return red
	default:
		return ""
	}
}
