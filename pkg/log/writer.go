package log

type Writer func(log *logger, level Level, msg string, extraFields ...Field)
