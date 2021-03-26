package time

import "time"

type DateTime struct {
	time.Time
}

func (t *DateTime) UnmarshalString(s string) error {
	n, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	t.Time = n
	return nil
}

func (t *DateTime) MarshalString() (string, error) {
	return t.Format(time.RFC3339Nano), nil
}

func Now() DateTime {
	return DateTime{time.Now().UTC()}
}
