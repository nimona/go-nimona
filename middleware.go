package fabric

type Middleware interface {
	Handler
	Negotiator
}
