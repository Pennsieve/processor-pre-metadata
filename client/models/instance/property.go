package instance

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/client/models/datatypes"
)

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
	var simpleType datatypes.SimpleType
	if err := json.Unmarshal(p.DataType, &simpleType); err != nil {
		var arrayType datatypes.ArrayDataType
		if arrErr := json.Unmarshal(p.DataType, &arrayType); arrErr != nil {
			return nil, fmt.Errorf("data type %s is not simple or array: %w", p.DataType, errors.Join(err, arrErr))
		}
		return arrayType, nil
	}
	return simpleType, nil
}

func (p Property) LongValue() (int64, error) {
	dataType, err := p.DecodeDataType()
	if err != nil {
		return 0, err
	}
	switch dt := dataType.(type) {
	case datatypes.SimpleType:
		if dt == datatypes.LongType {
			return int64(p.Value.(float64)), nil
		} else {
			return 0, fmt.Errorf("data type is not Long: %s", dt)
		}
	case datatypes.ArrayDataType:
		return 0, fmt.Errorf("data type is not Long: %T", dt)
	default:
		return 0, fmt.Errorf("unknown dataType: %T", dt)
	}

}

func (p Property) ArrayValue() (any, error) {
	dataType, err := p.DecodeDataType()
	if err != nil {
		return nil, err
	}
	switch dt := dataType.(type) {
	case datatypes.ArrayDataType:
		if dt.Type == datatypes.ArrayType && dt.Items.Type == datatypes.LongType {
			var longs []int64
			if p.Value == nil {
				return longs, nil
			}
			for _, l := range p.Value.([]any) {
				longs = append(longs, int64(l.(float64)))
			}
			return longs, nil
		} else if dt.Type == datatypes.ArrayType && dt.Items.Type == datatypes.StringType {
			return convertArray[string](p.Value), nil
		} else if dt.Type == datatypes.ArrayType && dt.Items.Type == datatypes.DoubleType {
			return convertArray[float64](p.Value), nil
		} else if dt.Type == datatypes.ArrayType && dt.Items.Type == datatypes.BooleanType {
			return convertArray[bool](p.Value), nil
		} else if dt.Type == datatypes.ArrayType && dt.Items.Type == datatypes.DateType {
			return convertArray[string](p.Value), nil
		} else {
			return nil, fmt.Errorf("data type is not array of Longs: %s of %s", dt.Type, dt.Items.Type)
		}
	case datatypes.SimpleType:
		return nil, fmt.Errorf("data type is not array: %T", dt)
	default:
		return nil, fmt.Errorf("unknown dataType: %T", dt)
	}
}

func convertArray[T any](src any) []T {
	if src == nil {
		return []T(nil)
	}
	srcSlice := src.([]any)
	converted := make([]T, len(srcSlice))
	for i, v := range srcSlice {
		converted[i] = v.(T)
	}
	return converted
}
