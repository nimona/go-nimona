package fabric

// middleware is a definition of  what a middleware is,
// take in one handlerfunc and wrap it within another handlerfunc
type middleware func(HandlerFunc) HandlerFunc

// BuildChain builds the middlware chain recursively, functions are first class
func BuildChain(f HandlerFunc, m ...Middleware) HandlerFunc {
	// if our chain is done, use the original handlerfunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the handlerfuncs
	return m[0].Wrap(BuildChain(f, m[1:cap(m)]...))
}
