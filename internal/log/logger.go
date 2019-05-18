package log

import (
	"encoding/json"
	"fmt"

	"nimona.io/internal/context"
)

type (
	logger struct {
		name    string
		context context.Context
		parent  *logger
		fields  []Field
	}
	Field struct {
		Key   string
		Value interface{}
	}
	Logger interface {
		With(fields ...Field) *logger
		Named(name string) *logger
		Debug(msg string, fields ...Field)
		Info(msg string, fields ...Field)
		Warn(msg string, fields ...Field)
		Error(msg string, fields ...Field)
		Panic(msg string, fields ...Field)
		Fatal(msg string, fields ...Field)
	}
)

var (
	DefaultLogger = &logger{}
)

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

func FromContext(ctx context.Context) *logger {
	log := &logger{
		context: ctx,
	}
	return log
}
func (log *logger) write(level Level, msg string, extraFields ...Field) {
	ctx := log.getContext()
	fields := log.getFields()
	fields = append(fields, extraFields...)

	res := map[string]interface{}{}
	cID := context.GetCorrelationID(ctx)
	if cID == "" {
		cID = "-"
	}

	for _, field := range fields {
		k := field.Key
		v := field.Value
		if s, ok := v.(interface{ String() string }); ok {
			v = s.String()
		} else if s, ok := v.(interface{ Error() string }); ok {
			v = s.Error()
		}
		if _, ok := res[k]; !ok {
			res[k] = v
		}
	}

	j, _ := json.Marshal(res)
	fmt.Printf("%s level=%d message=%s fields=%s\n", cID, level, msg, string(j))
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

func (log *logger) With(fields ...Field) *logger {
	nlog := &logger{
		parent: log,
		fields: fields,
	}
	return nlog
}

func (log *logger) Named(name string) *logger {
	nlog := &logger{
		name:   name,
		parent: log,
	}
	return nlog
}

func (log *logger) Debug(msg string, fields ...Field) {
	log.write(DebugLevel, msg, fields...)
}

func (log *logger) Info(msg string, fields ...Field) {
	log.write(InfoLevel, msg, fields...)
}

func (log *logger) Warn(msg string, fields ...Field) {
	log.write(WarnLevel, msg, fields...)
}

func (log *logger) Error(msg string, fields ...Field) {
	log.write(ErrorLevel, msg, fields...)
}

func (log *logger) Panic(msg string, fields ...Field) {
	log.write(PanicLevel, msg, fields...)
}

func (log *logger) Fatal(msg string, fields ...Field) {
	log.write(FatalLevel, msg, fields...)
}
