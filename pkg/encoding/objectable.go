package encoding

// TODO might just be the worst interface name ever
type objectable interface {
	ToObject() *Object
}
