package mkcontainer

type Item any

type ItemWithGUID interface {
	GetGUID() string
}

type ItemWithName interface {
	GetName() string
}
