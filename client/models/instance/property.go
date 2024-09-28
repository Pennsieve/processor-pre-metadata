package instance

import (
	"encoding/json"
	"fmt"
)

type SimpleType string

const StringType SimpleType = "String"
const LongType SimpleType = "Long"
const DoubleType SimpleType = "Double"
const BooleanType SimpleType = "Boolean"
const DateType SimpleType = "Date"

type ComplexType string

const ArrayType ComplexType = "array"

type ArrayDataType struct {
	Type ComplexType `json:"type"`
}

type Property struct {
	ConceptTitle bool `json:"conceptTitle"`
	// DataType can be a string or a JSON object
	DataType    json.RawMessage `json:"dataType"`
	Default     bool            `json:"default"`
	DisplayName string          `json:"displayName"`
	Locked      bool            `json:"locked"`
	Name        string          `json:"name"`
	Required    bool            `json:"required"`
	Value       any             `json:"value"`
}

func (p Property) DecodeDataType() (any, error) {
	var simpleType SimpleType
	if err := json.Unmarshal(p.DataType, &simpleType); err != nil {
		return nil, fmt.Errorf("error unmarshalling DataType: %w", err)
	}
	return simpleType, nil
}

func (p Property) DecodeValue() (any, error) {
	dataType, err := p.DecodeDataType()
	if err != nil {
		return nil, err
	}

	switch dt := dataType.(type) {
	case SimpleType:
		if dt == LongType {
			return int64(p.Value.(float64)), nil
		} else {
			return p.Value, nil
		}
	case ArrayDataType:
		return p.Value, nil
	default:
		return nil, fmt.Errorf("unknown dataType: %T", dt)
	}
}
