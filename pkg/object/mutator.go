package object

// Mutator receives an object, mutates it, or errors
type Mutator interface {
	Mutate(o *Object) error
}
