package object

// TODO might just be the worst interface name ever
type objectable interface {
	ToObject() *Object
}

// Objectable is a temp interface, we should consider removing
type Objectable interface {
	objectable
}
