package log

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/muesli/termenv"

	"nimona.io/pkg/context"
)

var (
	p        = termenv.ColorProfile()
	bgColors = map[Level]termenv.Color{
		DebugLevel: p.Color("#D290E4"),
		InfoLevel:  p.Color("#A8CC8C"),
		WarnLevel:  p.Color("#DBAB79"),
		ErrorLevel: p.Color("#E88388"),
		PanicLevel: p.Color("#E88388"),
		FatalLevel: p.Color("#E88388"),
	}
)

func StringWriter() Writer {
	return func(log *logger, level Level, msg string, extraFields ...Field) {
		ctx := log.getContext()
		fields := log.getFields()
		fields = append(fields, extraFields...)
		res := map[string]interface{}{}
		dt := ""
		for _, field := range fields {
			k := field.Key
			v := field.Value
			if v == nil {
				continue
			}
			if k == "datetime" {
				dt = v.(string)
				continue
			}
			if strings.HasPrefix(k, "build.") {
				continue
			}
			// nolint: gocritic
			if s, ok := v.(interface{ String() string }); ok {
				v = s.String()
			} else if s, ok := v.(interface{ Error() string }); ok {
				v = s.Error()
			}
			if _, ok := res[k]; !ok {
				res[k] = v
			}
		}

		meta := ""
		if ok, _ := strconv.ParseBool(os.Getenv("NIMONA_LOG_NOMETA")); !ok {
			if context.GetCorrelationID(ctx) != "" {
				res["ctx"] = context.GetCorrelationID(ctx)
			}
			b, _ := yaml.Marshal(res)
			s := strings.ReplaceAll("\n"+string(b), "\n", "\n\t\t     ")
			meta = termenv.String(s).Faint().String()
		}
		fmt.Fprintf(
			log.output,
			"%s %s %s %s\n",
			dt[11:23],
			termenv.
				String(" "+fmt.Sprintf("%- 6s", levels[level])).
				Foreground(p.Color("0")).
				Background(
					bgColors[level],
				),
			termenv.
				String(msg).
				Foreground(
					bgColors[level],
				),
			meta,
		)
	}
}
