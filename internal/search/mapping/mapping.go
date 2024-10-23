package mapping

type FieldType string

const (
	FieldTypeKeyword  FieldType = "keyword"
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeShingles FieldType = "shingles"
)
