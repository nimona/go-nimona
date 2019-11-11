package crypto

func NewCertificate(k PrivateKey) *Certificate {
	c := &Certificate{
		Subject: k.PublicKey(),
	}
	return c
}

func (c *Certificate) Sign(k PrivateKey) {
	s, _ := NewSignature(k, c.ToObject())
	c.Signature = s
}
