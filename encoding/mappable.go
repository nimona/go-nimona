package encoding

type Mappable interface {
	ObjectMap() map[string]interface{}
}
