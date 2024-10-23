package searchterms

type CompOp string

const (
	CompOpEq        CompOp = "="
	CompOpNeq       CompOp = "!="
	CompOpLike      CompOp = "~="
	CompOpFuzzyLike CompOp = "~"
	CompOpLt        CompOp = "<"
	CompOpLe        CompOp = "<="
	CompOpGt        CompOp = ">"
	CompOpGe        CompOp = ">="
)
