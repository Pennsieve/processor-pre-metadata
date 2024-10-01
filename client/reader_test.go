package client

import (
	"github.com/pennsieve/processor-pre-metadata/client/models/instance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewReader(t *testing.T) {
	reader, err := NewReader("testdata")
	require.NoError(t, err)
	assert.Len(t, reader.ModelNamesToSchemaElements, 3)
	assert.Equal(t, "83964537-46d2-4fb5-9408-0b6262a42a56", reader.ModelNamesToSchemaElements["location"].ID)
	assert.Equal(t, "bb04a8ce-03c9-4801-a0d9-e35cea53ac1b", reader.ModelNamesToSchemaElements["object"].ID)
	assert.Equal(t, "7931cbe6-7494-4c0b-95f0-9f4b34edc73b", reader.ModelNamesToSchemaElements["subject"].ID)
}

func TestReader_GetRecordsForModel(t *testing.T) {
	reader, err := NewReader("testdata")
	require.NoError(t, err)

	records, err := reader.GetRecordsForModel("Object")
	require.NoError(t, err)
	assert.Len(t, records, 3)

	idToPropNameToProp := map[string]map[string]instance.Property{}
	for _, r := range records {
		propNameToProp := map[string]instance.Property{}
		for _, p := range r.Values {
			propNameToProp[p.Name] = p
		}
		idToPropNameToProp[r.ID] = propNameToProp
		assert.Len(t, r.Values, 7)
		assert.Equal(t, "object", r.Type)
	}
	assert.Len(t, idToPropNameToProp, 3)

	// A record with some null property values
	{
		propNameToProp := idToPropNameToProp["5b07e038-9829-46c9-b698-bf4efef81341"]
		assert.Len(t, propNameToProp, 7)

		// Name
		name := propNameToProp["name"]
		assertSimpleType(t, instance.StringType, "stone", name)

		// ID
		id := propNameToProp["id"]
		assertSimpleType(t, instance.LongType, int64(1), id)

		// Weights
		weights := propNameToProp["weights"]
		assertArrayType(t, instance.ArrayDataType{
			Type:  instance.ArrayType,
			Items: instance.ItemsType{Type: instance.LongType},
		}, []int64(nil), weights)

		// Synonyms
		synonyms := propNameToProp["synonyms"]
		assertArrayType(t, instance.ArrayDataType{
			Type:  instance.ArrayType,
			Items: instance.ItemsType{Type: instance.StringType},
		}, []string(nil), synonyms)

		// GPA
		gpa := propNameToProp["gpa"]
		assertSimpleType(t, instance.DoubleType, nil, gpa)

		// Birthday
		birthday := propNameToProp["birthday"]
		assertSimpleType(t, instance.DateType, nil, birthday)

		// IsSolid
		isSolid := propNameToProp["is_solid"]
		assertSimpleType(t, instance.BooleanType, nil, isSolid)
	}

	// A record with no null property values
	{
		propNameToProp := idToPropNameToProp["a9b9d03b-19b3-4a43-b40e-5673ec955e49"]
		assert.Len(t, propNameToProp, 7)

		// Name
		name := propNameToProp["name"]
		assertSimpleType(t, instance.StringType, "whatsit", name)

		// ID
		id := propNameToProp["id"]
		assertSimpleType(t, instance.LongType, int64(57), id)

		// Weights
		weights := propNameToProp["weights"]
		assertArrayType(t, instance.ArrayDataType{
			Type:  instance.ArrayType,
			Items: instance.ItemsType{Type: instance.LongType},
		}, []int64{3, 5, 7}, weights)

		// Synonyms
		synonyms := propNameToProp["synonyms"]
		assertArrayType(t, instance.ArrayDataType{
			Type:  instance.ArrayType,
			Items: instance.ItemsType{Type: instance.StringType},
		}, []string{"thingamabob", "whosit", "doo-dad"}, synonyms)

		// GPA
		gpa := propNameToProp["gpa"]
		assertSimpleType(t, instance.DoubleType, 6.78, gpa)

		// Birthday
		birthday := propNameToProp["birthday"]
		require.NoError(t, err)
		assertSimpleType(t, instance.DateType, "2024-09-26T22:01:04", birthday)

		// IsSolid
		isSolid := propNameToProp["is_solid"]
		assertSimpleType(t, instance.BooleanType, "true", isSolid)
	}

}

func assertSimpleType(t *testing.T, expectedType instance.SimpleType, expectedValue any, actualProperty instance.Property) bool {
	dataType, err := actualProperty.DecodeDataType()
	if !assert.NoError(t, err) {
		return false
	}
	if !assert.Equal(t, expectedType, dataType) {
		return false
	}

	actualValue := actualProperty.Value
	if expectedType == instance.LongType {
		actualValue, err = actualProperty.LongValue()
		if !assert.NoError(t, err) {
			return false
		}
	}

	if !assert.Equal(t, expectedValue, actualValue) {
		return false
	}
	return true
}

func assertArrayType(t *testing.T, expectedType instance.ArrayDataType, expectedValue any, actualProperty instance.Property) bool {
	dataType, err := actualProperty.DecodeDataType()
	if !assert.NoError(t, err) {
		return false
	}
	if !assert.IsType(t, instance.ArrayDataType{}, dataType) {
		return false
	}
	actualDataType := dataType.(instance.ArrayDataType)
	if !assert.Equal(t, expectedType.Type, actualDataType.Type) {
		return false
	}
	if !assert.Equal(t, expectedType.Items.Type, actualDataType.Items.Type) {
		return false
	}

	actualValue, err := actualProperty.ArrayValue()
	if !assert.NoError(t, err) {
		return false
	}

	if !assert.Equal(t, expectedValue, actualValue) {
		return false
	}
	return true
}
