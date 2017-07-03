package main

type A struct{}

/*
string pointerify
*/
type Ptr struct {
	AA []A
}

type Base struct {
}

// Foo is a struct
type Foo struct {
	Boo  bool `json:"boo,omitempty" yaml:"booY"`
	Zoo  bool // Specific comment
	SPtr *Ptr
	VPtr Ptr
	APtr []Ptr // should become a pointer
	MPtr map[string]Ptr

	Base
}

func main() {}
