package crypto

type Fingerprint string

func (f Fingerprint) String() string {
	return string(f)
}

func (f Fingerprint) Address() string {
	return "peer:" + string(f)
}
