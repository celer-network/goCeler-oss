// Copyright 2018 Celer Network

// API:
// Trace[f|ln], Debug[f|ln], Info[f|ln], Warn[f|ln], Error[f|ln], Fatal[f|ln], Panic[f|ln],
// SetLevel(Level), IsLevelEnabled(Level)

// Use flags to change standard log behavior. flag.Parse() must be called.
// -loglevel=[trace|debug|info|warn|error|fatal|panic] default: info
//		Logs at or above this threshold.
// -logdbginclude=[string]
//		include trace and debug level log for specified file paths
//		Can be set multiple times to allow multiple filters.
// -logdbgexclude=[string]
//		exclude trace and debug level log for specified file paths
//		Can be set multiple times to allow multiple filters.
// -logdir=[string] default: ""
//		If non-empty, write log files in this directory instead of stderr.
// -logname=[string] default: main file name
//		Log file name, log creation date and time will be appended automatically.
// -logprefix=[string] default: ""
//		Sets the output prefix for the standard logger.
// -logrotate=[true|false] default: true
//		Daily rotate log file
// -logcolor=[true|false] default: false
//		Use color-coded log level at terminal, should be false if logdir is non-empty
// -loglocaltime=[true|false] default: false
//		Use the local time zone rather than UTC.
// -loglongfile=[true|false] default: false
//		Show long file path (e.g., goCeler-oss/clog/test/clogtest.go:17).
// -logpathsplit=[string] default: "celer-network/"
//		Long file name starts after the splittter, if -loglongfile=true

// Example log format: "2006-01-02 15:04:05 |DEBUG| test.go:9: log message"

package clog

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Level of severity.
type Level uint

// Log levels.
const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

const (
	black = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

type levelmeta struct {
	print string
	name  string
	color int
}

var levelInfo = []levelmeta{
	TraceLevel: levelmeta{"|TRACE| ", "trace", blue},
	DebugLevel: levelmeta{"|DEBUG| ", "debug", cyan},
	InfoLevel:  levelmeta{"|INFO | ", "info", green},
	WarnLevel:  levelmeta{"|WARN | ", "warn", yellow},
	ErrorLevel: levelmeta{"|ERROR| ", "error", red},
	FatalLevel: levelmeta{"|FATAL| ", "fatal", red},
	PanicLevel: levelmeta{"|PANIC| ", "panic", red},
}

type arrayFlags []string

// Logger obejct
type Logger struct {
	file       *os.File     // log file writer
	filetime   time.Time    // file created time
	mu         sync.Mutex   // protect log writer
	dir        string       // write log into directory instead of stderr
	name       string       // log file name
	prefix     string       // log entry prefix
	level      Level        // log level
	dbginclude arrayFlags   // include trace and debug level log for specified file paths
	dbgexclude arrayFlags   // exclude trace and debug level log for specified file paths
	rotate     bool         // daily rotate log file
	color      bool         // enable terminal color
	localtime  bool         // use local time
	longfile   bool         // show long file path
	pathsplit  string       // file path splitter
	rw         sync.RWMutex // protect config values
}

var std Logger

var bufferPool *sync.Pool // log msg buffer pool

func init() {
	flag.Var(&std.level, "loglevel", "log at or above this level (options: trace/debug/info/warn/error/fatal/panic)")
	flag.Var(&std.dbginclude, "logdbginclude", "include trace and debug level log for specified file paths")
	flag.Var(&std.dbgexclude, "logdbgexclude", "exclude trace and debug level log for specified file paths")
	flag.StringVar(&std.dir, "logdir", "", "if non-empty, write log files in this directory instead of stderr")
	flag.StringVar(&std.name, "logname", filepath.Base(os.Args[0]), "Log file name, log creation date and time will be appended")
	flag.StringVar(&std.prefix, "logprefix", "", "log entry prefix")
	flag.BoolVar(&std.rotate, "logrotate", true, "daily rotate log file")
	flag.BoolVar(&std.color, "logcolor", false, "use color-coded log level")
	flag.BoolVar(&std.localtime, "loglocaltime", false, "use the local time zone rather than UTC")
	flag.BoolVar(&std.longfile, "loglongfile", false, "show long file path")
	flag.StringVar(&std.pathsplit, "logpathsplit", "celer-network/", "file path splitter")

	bufferPool = &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	std.level = InfoLevel
	std.filetime = time.Time{}
}

var onceLogFile sync.Once

func (l *Logger) output(msg string, level Level) {
	// data and time
	var t time.Time
	if l.localtime {
		t = time.Now()
	} else {
		t = time.Now().UTC()
	}

	// get a buffer from the pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// format header
	if l.prefix != "" {
		buf.WriteString(l.prefix + " ")
	}
	buf.WriteString(t.Format("2006-01-02 15:04:05.000 "))

	// log level
	if l.color {
		fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m", levelInfo[level].color, levelInfo[level].print)
	} else {
		buf.WriteString(levelInfo[level].print)
	}

	// file name and line number
	_, file, line, ok := runtime.Caller(2)
	if level <= DebugLevel {
		if len(std.dbgexclude) > 0 {
			for _, path := range std.dbgexclude {
				if strings.Contains(file, path) {
					return
				}
			}
		}
		if len(std.dbginclude) > 0 {
			pass := false
			for _, path := range std.dbginclude {
				if strings.Contains(file, path) {
					pass = true
					break
				}
			}
			if !pass {
				return
			}
		}
	}
	if !ok {
		file = "???"
		line = 0
	} else if l.longfile {
		split := strings.Index(file, l.pathsplit)
		if split >= 0 {
			file = file[split+len(l.pathsplit):]
		}
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	buf.WriteString(file)
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(line))
	buf.WriteString(": ")

	// log message
	buf.WriteString(msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		buf.WriteByte('\n')
	}

	// write to output
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.dir != "" {
		onceLogFile.Do(func() { // create folder and first log file
			err := os.MkdirAll(l.dir, 0755)
			if err != nil {
				os.Stderr.Write(buf.Bytes())
				l.exit(err)
			}
			err = l.createFile(t)
			if err != nil {
				os.Stderr.Write(buf.Bytes())
				l.exit(err)
			}
		})
		if l.rotate && t.Day() != l.filetime.Day() {
			err := l.rotateFile(t)
			if err != nil {
				os.Stderr.Write(buf.Bytes())
				l.exit(err)
			}
		}
		l.file.Write(buf.Bytes())
	} else {
		os.Stderr.Write(buf.Bytes())
	}
}

func (l *Logger) exit(err error) {
	fmt.Fprintf(os.Stderr, "FATAL: log exiting because of error: %s\n", err)
	os.Exit(2)
}

// close the file if it is open
func (l *Logger) close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	if err != nil {
		return fmt.Errorf("can't close logfile: %s", err)
	}
	l.file = nil
	return err
}

func (l *Logger) rotateFile(t time.Time) error {
	if err := l.close(); err != nil {
		return err
	}
	if err := l.createFile(t); err != nil {
		return err
	}
	return nil
}

// create a new log file
func (l *Logger) createFile(t time.Time) error {
	timezone, _ := t.Zone()
	filename := fmt.Sprintf("%s.%04d.%02d.%02d.%02dh%02dm%02ds%s.log",
		l.name,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		timezone)
	fname := filepath.Join(l.dir, filename)
	var err error
	l.file, err = os.Create(fname)
	if err == nil {
		logpath, _ := filepath.Abs(fname)
		os.Stderr.Write([]byte("Log to " + logpath + "\n"))
		l.filetime = t
		return nil
	}
	return fmt.Errorf("log: cannot create log file: %s", err)
}

func (l *Logger) isLevelEnabled(level Level) bool {
	l.rw.RLock()
	defer l.rw.RUnlock()
	return l.level <= level
}

// API for the standard logger

func SetLevel(level Level) {
	std.rw.Lock()
	defer std.rw.Unlock()
	if level > PanicLevel {
		level = PanicLevel
	}
	std.level = level
}

func IsLevelEnabled(level Level) bool {
	return std.isLevelEnabled(level)
}

func Trace(v ...interface{}) {
	if std.isLevelEnabled(TraceLevel) {
		std.output(fmt.Sprint(v...), TraceLevel)
	}
}

func Tracef(format string, v ...interface{}) {
	if std.isLevelEnabled(TraceLevel) {
		std.output(fmt.Sprintf(format, v...), TraceLevel)
	}
}

func Traceln(v ...interface{}) {
	if std.isLevelEnabled(TraceLevel) {
		std.output(fmt.Sprintln(v...), TraceLevel)
	}
}

func Debug(v ...interface{}) {
	if std.isLevelEnabled(DebugLevel) {
		std.output(fmt.Sprint(v...), DebugLevel)
	}
}

func Debugf(format string, v ...interface{}) {
	if std.isLevelEnabled(DebugLevel) {
		std.output(fmt.Sprintf(format, v...), DebugLevel)
	}
}

func Debugln(v ...interface{}) {
	if std.isLevelEnabled(DebugLevel) {
		std.output(fmt.Sprintln(v...), DebugLevel)
	}
}

func Info(v ...interface{}) {
	if std.isLevelEnabled(InfoLevel) {
		std.output(fmt.Sprint(v...), InfoLevel)
	}
}

func Infof(format string, v ...interface{}) {
	if std.isLevelEnabled(InfoLevel) {
		std.output(fmt.Sprintf(format, v...), InfoLevel)
	}
}

func Infoln(v ...interface{}) {
	if std.isLevelEnabled(InfoLevel) {
		std.output(fmt.Sprintln(v...), InfoLevel)
	}
}

func Warn(v ...interface{}) {
	if std.isLevelEnabled(WarnLevel) {
		std.output(fmt.Sprint(v...), WarnLevel)
	}
}

func Warnf(format string, v ...interface{}) {
	if std.isLevelEnabled(WarnLevel) {
		std.output(fmt.Sprintf(format, v...), WarnLevel)
	}
}

func Warnln(v ...interface{}) {
	if std.isLevelEnabled(WarnLevel) {
		std.output(fmt.Sprintln(v...), WarnLevel)
	}
}

func Error(v ...interface{}) {
	if std.isLevelEnabled(ErrorLevel) {
		std.output(fmt.Sprint(v...), ErrorLevel)
	}
}

func Errorf(format string, v ...interface{}) {
	if std.isLevelEnabled(ErrorLevel) {
		std.output(fmt.Sprintf(format, v...), ErrorLevel)
	}
}

func Errorln(v ...interface{}) {
	if std.isLevelEnabled(ErrorLevel) {
		std.output(fmt.Sprintln(v...), ErrorLevel)
	}
}

func Fatal(v ...interface{}) {
	if std.isLevelEnabled(FatalLevel) {
		std.output(fmt.Sprint(v...), FatalLevel)
		os.Exit(1)
	}
}

func Fatalf(format string, v ...interface{}) {
	if std.isLevelEnabled(FatalLevel) {
		std.output(fmt.Sprintf(format, v...), FatalLevel)
		os.Exit(1)
	}
}

func Fatalln(v ...interface{}) {
	if std.isLevelEnabled(FatalLevel) {
		std.output(fmt.Sprintln(v...), FatalLevel)
		os.Exit(1)
	}
}

func Panic(v ...interface{}) {
	if std.isLevelEnabled(PanicLevel) {
		s := fmt.Sprint(v...)
		std.output(s, PanicLevel)
		panic(s)
	}
}

func Panicf(format string, v ...interface{}) {
	if std.isLevelEnabled(PanicLevel) {
		s := fmt.Sprintf(format, v...)
		std.output(s, PanicLevel)
		panic(s)
	}
}

func Panicln(v ...interface{}) {
	if std.isLevelEnabled(PanicLevel) {
		s := fmt.Sprintln(v...)
		std.output(s, PanicLevel)
		panic(s)
	}
}

var levelSetByFlag = false

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	var valid bool
	value = strings.ToLower(value)
	for i, info := range levelInfo {
		if info.name == value {
			SetLevel(Level(i))
			valid = true
			levelSetByFlag = true
			break
		}
	}
	if !valid {
		os.Stderr.Write([]byte("Error: invalid log level flag, use default InfoLevel\n"))
		SetLevel(InfoLevel)
	}
	return nil
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

// LevelSetByFlag returns true if log level is set by flag,
// so that it is possible to let flag input can override SetLevel() from config file
func LevelSetByFlag() bool {
	return levelSetByFlag
}

// Set is part of the flag.Value interface.
func (f *arrayFlags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *arrayFlags) String() string {
	return strings.Join(*f, " ")
}
