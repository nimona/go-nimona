package fixtures

type Composite struct {
	Foo string
}

// String

func (c *Composite) UnmarshalString(s string) error {
	c.Foo = s
	return nil
}

func (c Composite) MarshalString() (string, error) {
	return c.Foo, nil
}

// Bytes

func (c *Composite) UnmarshalBytes(s []byte) error {
	c.Foo = string(s)
	return nil
}

func (c Composite) MarshalBytes() ([]byte, error) {
	return []byte(c.Foo), nil
}
