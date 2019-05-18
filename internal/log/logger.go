package log

import (
	"encoding/json"
	"fmt"
	"reflect"

	"go.uber.org/zap"

	"nimona.io/internal/context"
)

type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

type (
	logger struct {
		name    string
		context context.Context
		parent  *logger
		fields  []interface{}
	}
	StringField struct {
		Key   string
		Value string
	}
	ZapLogger interface {
		With(fields ...interface{}) *logger
		Named(name string) *logger
		Debug(msg string, fields ...interface{})
		Info(msg string, fields ...interface{})
		Warn(msg string, fields ...interface{})
		Error(msg string, fields ...interface{})
		Panic(msg string, fields ...interface{})
		Fatal(msg string, fields ...interface{})
	}
)

var (
	DefaultLogger = &logger{}
)

func String(k, v string) StringField {
	return StringField{
		Key:   k,
		Value: v,
	}
}

func Logger(ctx context.Context) *logger {
	log := &logger{
		context: ctx,
	}
	return log
}
func (log *logger) write(level Level, msg string, extraFields ...interface{}) {
	ctx := log.getContext()
	fields := log.getFields()
	fields = append(fields, extraFields...)

	res := map[string]interface{}{}
	cID := context.GetCorrelationID(ctx)
	if cID == "" {
		cID = "-"
	}

	for _, field := range fields {
		var (
			k string
			v interface{}
		)
		switch f := field.(type) {
		case StringField:
			k = f.Key
			v = f.Value
		case zap.Field:
			k = f.Key
			if f.Interface != nil {
				v = f.Interface
			} else if f.String != "" {
				v = f.String
			} else {
				v = f.Integer
			}
		default:
			fmt.Println("___", f, reflect.TypeOf(field))
		}
		if k == "" {
			continue
		}
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

func (log *logger) getFields() []interface{} {
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

func (log *logger) With(fields ...interface{}) *logger {
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

func (log *logger) Debug(msg string, fields ...interface{}) {
	log.write(DebugLevel, msg, fields...)
}

func (log *logger) Info(msg string, fields ...interface{}) {
	log.write(InfoLevel, msg, fields...)
}

func (log *logger) Warn(msg string, fields ...interface{}) {
	log.write(WarnLevel, msg, fields...)
}

func (log *logger) Error(msg string, fields ...interface{}) {
	log.write(ErrorLevel, msg, fields...)
}

func (log *logger) Panic(msg string, fields ...interface{}) {
	log.write(PanicLevel, msg, fields...)
}

func (log *logger) Fatal(msg string, fields ...interface{}) {
	log.write(FatalLevel, msg, fields...)
}
