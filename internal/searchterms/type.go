package searchterms

type Type string

const (
	IntType      Type = "int"
	StringType   Type = "string"
	DurationType Type = "duration"
)

func (t Type) Kind() Type {
	return t
}

func (t Type) Equal(t2 Type) bool {
	return t == t2
}

func (t Type) String() string {
	return string(t)
}
