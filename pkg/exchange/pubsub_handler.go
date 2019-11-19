package exchange

func HandleEnvelopeSubscription(
	sub EnvelopeSubscription,
	handler func(*Envelope) error,
) {
	for {
		e, err := sub.Next()
		if err != nil {
			return
		}
		if err := handler(e); err != nil {
			return
		}
	}
}
