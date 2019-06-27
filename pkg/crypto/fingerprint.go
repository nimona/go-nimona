package crypto

type Fingerprint string

func (f Fingerprint) String() string {
	return string(f)
}
