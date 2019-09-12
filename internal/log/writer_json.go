package log

import (
	"encoding/json"
	"fmt"

	"nimona.io/internal/context"
)

func JSONWriter() Writer {
	return func(log *logger, level Level, msg string, extraFields ...Field) {
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

		m := map[string]interface{}{
			"cID":    cID,
			"fields": fields,
			"msg":    msg,
			"level":  levels[level],
		}

		b, _ := json.Marshal(m)
		fmt.Println(string(b))
	}
}
