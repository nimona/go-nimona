package log

import (
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"nimona.io/pkg/context"
)

type (
	logger struct {
		name    string
		context context.Context
		parent  *logger
		fields  []Field
		output  io.Writer
		writer  Writer
	}
	Field struct {
		Key   string
		Value interface{}
	}
	Logger interface {
		With(fields ...Field) Logger
		Named(name string) Logger
		Debug(msg string, fields ...Field)
		Info(msg string, fields ...Field)
		Warn(msg string, fields ...Field)
		Error(msg string, fields ...Field)
		Panic(msg string, fields ...Field)
		Fatal(msg string, fields ...Field)
	}
)

var (
	DefaultLogLevel Level     = ErrorLevel
	DefaultOutput   io.Writer = os.Stdout
	DefaultWriter   Writer    = JSONWriter()
	DefaultLogger   Logger    = &logger{
		writer: DefaultWriter,
		output: DefaultOutput,
	}
)

// TODO is there an alternative to init for here?
// nolint: gochecknoinits
func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		DefaultLogLevel = DebugLevel
	case "INFO":
		DefaultLogLevel = InfoLevel
	case "WARN", "WARNING":
		DefaultLogLevel = WarnLevel
	case "ERR", "ERROR":
		DefaultLogLevel = ErrorLevel
	case "PANIC":
		DefaultLogLevel = PanicLevel
	case "FATAL":
		DefaultLogLevel = FatalLevel
	default:
		DefaultLogLevel = ErrorLevel
	}
}

func Stack() Field {
	return Field{
		Key:   "stack",
		Value: string(debug.Stack()),
	}
}

func String(k, v string) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

func Strings(k string, v []string) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

func Int(k string, v int) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

func Bool(k string, v bool) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

func Error(v error) Field {
	return Field{
		Key:   "error",
		Value: v,
	}
}

func Any(k string, v interface{}) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

func FromContext(ctx context.Context) Logger {
	log := &logger{
		parent:  DefaultLogger.(*logger),
		context: ctx,
		writer:  DefaultWriter,
		output:  DefaultOutput,
	}
	return log
}

func (log *logger) write(level Level, msg string, extraFields ...Field) {
	extraFields = append(
		extraFields,
		String("$$time", time.Now().Format(time.RFC3339Nano)),
	)
	log.writer(log, level, msg, extraFields...)
}

func (log *logger) getFields() []Field {
	fields := log.fields
	if log.parent != nil {
		fields = append(fields, log.parent.getFields()...)
	}
	return fields
}

func (log *logger) getContext() context.Context {
	if log.context != nil {
		return log.context
	}

	if log.parent != nil {
		return log.parent.getContext()
	}

	return nil
}

func (log *logger) With(fields ...Field) Logger {
	nlog := &logger{
		parent: log,
		fields: fields,
		writer: log.writer,
		output: log.output,
	}
	return nlog
}

func (log *logger) Named(name string) Logger {
	nlog := &logger{
		name:   strings.Join([]string{log.name, name}, "/"),
		parent: log,
		writer: log.writer,
		output: log.output,
	}
	return nlog
}

func (log *logger) Debug(msg string, fields ...Field) {
	if DebugLevel >= DefaultLogLevel {
		log.write(DebugLevel, msg, fields...)
	}
}

func (log *logger) Info(msg string, fields ...Field) {
	if InfoLevel >= DefaultLogLevel {
		log.write(InfoLevel, msg, fields...)
	}
}

func (log *logger) Warn(msg string, fields ...Field) {
	if WarnLevel >= DefaultLogLevel {
		log.write(WarnLevel, msg, fields...)
	}
}

func (log *logger) Error(msg string, fields ...Field) {
	if ErrorLevel >= DefaultLogLevel {
		log.write(ErrorLevel, msg, fields...)
	}
}

func (log *logger) Panic(msg string, fields ...Field) {
	if PanicLevel >= DefaultLogLevel {
		log.write(PanicLevel, msg, fields...)
	}
}

func (log *logger) Fatal(msg string, fields ...Field) {
	if FatalLevel >= DefaultLogLevel {
		log.write(FatalLevel, msg, fields...)
	}
	os.Exit(1)
}
