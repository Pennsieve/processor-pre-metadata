package datatypes

type SimpleType string

const StringType SimpleType = "String"
const LongType SimpleType = "Long"
const DoubleType SimpleType = "Double"
const BooleanType SimpleType = "Boolean"
const DateType SimpleType = "Date"

type ComplexType string

const ArrayType ComplexType = "array"

type ArrayDataType struct {
	Type  ComplexType `json:"type"`
	Items ItemsType   `json:"items"`
}

type ItemsType struct {
	Type   SimpleType `json:"type"`
	Format string     `json:"format,omitempty"`
	Unit   string     `json:"unit,omitempty"`
}
