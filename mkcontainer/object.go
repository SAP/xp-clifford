package mkcontainer

type Object interface{}

type ObjectWithGUID interface {
	GetGUID() string
}

type ObjectWithName interface {
	GetName() string
}
