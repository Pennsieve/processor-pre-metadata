package client

import (
	"github.com/pennsieve/processor-pre-metadata/client/models/datatypes"
	"github.com/pennsieve/processor-pre-metadata/client/models/instance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewReader(t *testing.T) {
	reader, err := NewReader("testdata")
	require.NoError(t, err)

	assert.Equal(t, 3, reader.Schema.ModelCount())

	for modelName, expectedModelID := range map[string]string{
		"location": "83964537-46d2-4fb5-9408-0b6262a42a56",
		"object":   "bb04a8ce-03c9-4801-a0d9-e35cea53ac1b",
		"subject":  "7931cbe6-7494-4c0b-95f0-9f4b34edc73b",
	} {
		model, exists := reader.Schema.ModelByName(modelName)
		assert.True(t, exists)
		assert.Equal(t, expectedModelID, model.ID)
	}
	assert.Equal(t, 1, reader.Schema.LinkedPropertyCount())

	assert.NotNil(t, reader.Schema.Proxy())
	assert.Equal(t, "e18a8519-8368-4062-977a-60707c9c93ec", reader.Schema.Proxy().ID)
}

func TestReader_GetProxiesForModel(t *testing.T) {
	reader, err := NewReader("testdata")
	require.NoError(t, err)

	// location proxy instances
	{
		instancesByRecordID, err := reader.GetProxiesForModel("location")
		require.NoError(t, err)

		assert.Len(t, instancesByRecordID, 1)

		assert.Contains(t, instancesByRecordID, "e79e8d65-b094-4f36-94f2-1553cd84b4a2")
		instances := instancesByRecordID["e79e8d65-b094-4f36-94f2-1553cd84b4a2"]
		assert.Len(t, instances, 1)

		assert.Equal(t, "a6752c89-83d9-4191-8806-d55956e3217c", instances[0].ID)
		assert.Equal(t, "N:collection:e3c0abb8-7480-42af-9529-99cafe9ea235", instances[0].Content.NodeID)
	}

	//object proxy instances
	{
		instancesByRecordID, err := reader.GetProxiesForModel("object")
		require.NoError(t, err)

		assert.Len(t, instancesByRecordID, 2)

		assert.Contains(t, instancesByRecordID, "a9b9d03b-19b3-4a43-b40e-5673ec955e49")
		instances1 := instancesByRecordID["a9b9d03b-19b3-4a43-b40e-5673ec955e49"]
		assert.Len(t, instances1, 1)

		assert.Equal(t, "6baa77da-9760-4deb-8a19-c97c3286a259", instances1[0].ID)
		assert.Equal(t, "N:collection:95bb7c19-0e8e-42b2-b53f-f5ce7a08e42a", instances1[0].Content.NodeID)

		assert.Contains(t, instancesByRecordID, "bcf06e0c-42dc-4ce9-9c70-9ee6865ebc7c")
		instances2 := instancesByRecordID["bcf06e0c-42dc-4ce9-9c70-9ee6865ebc7c"]
		assert.Len(t, instances2, 1)

		assert.Equal(t, "15bebbdc-e479-462f-b094-043a29cecfc9", instances2[0].ID)
		assert.Equal(t, "N:package:f90ff4bc-e3e5-4a53-b545-158ea770fbd8", instances2[0].Content.NodeID)
	}

	//subject proxy instances
	{
		instancesByRecordID, err := reader.GetProxiesForModel("subject")
		require.NoError(t, err)

		assert.Empty(t, instancesByRecordID)
	}
}

func TestReader_GetRecordsForModel(t *testing.T) {
	reader, err := NewReader("testdata")
	require.NoError(t, err)

	records, err := reader.GetRecordsForModel("object")
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
		assertSimpleType(t, datatypes.StringType, "stone", name)

		// ID
		id := propNameToProp["id"]
		assertSimpleType(t, datatypes.LongType, int64(1), id)

		// Weights
		weights := propNameToProp["weights"]
		assertArrayType(t, datatypes.ArrayDataType{
			Type:  datatypes.ArrayType,
			Items: datatypes.ItemsType{Type: datatypes.LongType},
		}, []int64(nil), weights)

		// Synonyms
		synonyms := propNameToProp["synonyms"]
		assertArrayType(t, datatypes.ArrayDataType{
			Type:  datatypes.ArrayType,
			Items: datatypes.ItemsType{Type: datatypes.StringType},
		}, []string(nil), synonyms)

		// GPA
		gpa := propNameToProp["gpa"]
		assertSimpleType(t, datatypes.DoubleType, nil, gpa)

		// Birthday
		birthday := propNameToProp["birthday"]
		assertSimpleType(t, datatypes.DateType, nil, birthday)

		// IsSolid
		isSolid := propNameToProp["is_solid"]
		assertSimpleType(t, datatypes.BooleanType, nil, isSolid)
	}

	// A record with no null property values
	{
		propNameToProp := idToPropNameToProp["a9b9d03b-19b3-4a43-b40e-5673ec955e49"]
		assert.Len(t, propNameToProp, 7)

		// Name
		name := propNameToProp["name"]
		assertSimpleType(t, datatypes.StringType, "whatsit", name)

		// ID
		id := propNameToProp["id"]
		assertSimpleType(t, datatypes.LongType, int64(57), id)

		// Weights
		weights := propNameToProp["weights"]
		assertArrayType(t, datatypes.ArrayDataType{
			Type:  datatypes.ArrayType,
			Items: datatypes.ItemsType{Type: datatypes.LongType},
		}, []int64{3, 5, 7}, weights)

		// Synonyms
		synonyms := propNameToProp["synonyms"]
		assertArrayType(t, datatypes.ArrayDataType{
			Type:  datatypes.ArrayType,
			Items: datatypes.ItemsType{Type: datatypes.StringType},
		}, []string{"thingamabob", "whosit", "doo-dad"}, synonyms)

		// GPA
		gpa := propNameToProp["gpa"]
		assertSimpleType(t, datatypes.DoubleType, 6.78, gpa)

		// Birthday
		birthday := propNameToProp["birthday"]
		require.NoError(t, err)
		assertSimpleType(t, datatypes.DateType, "2024-09-26T22:01:04", birthday)

		// IsSolid
		isSolid := propNameToProp["is_solid"]
		assertSimpleType(t, datatypes.BooleanType, "true", isSolid)
	}

}

func assertSimpleType(t *testing.T, expectedType datatypes.SimpleType, expectedValue any, actualProperty instance.Property) bool {
	dataType, err := actualProperty.DecodeDataType()
	if !assert.NoError(t, err) {
		return false
	}
	if !assert.Equal(t, expectedType, dataType) {
		return false
	}

	actualValue := actualProperty.Value
	if expectedType == datatypes.LongType {
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

func assertArrayType(t *testing.T, expectedType datatypes.ArrayDataType, expectedValue any, actualProperty instance.Property) bool {
	dataType, err := actualProperty.DecodeDataType()
	if !assert.NoError(t, err) {
		return false
	}
	if !assert.IsType(t, datatypes.ArrayDataType{}, dataType) {
		return false
	}
	actualDataType := dataType.(datatypes.ArrayDataType)
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

func TestReader_GetLinkInstancesForProperty(t *testing.T) {
	reader, err := NewReader("testdata")
	require.NoError(t, err)

	links, err := reader.GetLinkInstancesForProperty("address")
	require.NoError(t, err)
	assert.Len(t, links, 1)

	linkInstance := links[0]
	assert.Equal(t, "address", linkInstance.Name)
	assert.Equal(t, "Address", linkInstance.DisplayName)

	assert.Equal(t, "address", linkInstance.Type)

	assert.Equal(t, "b7bcfc2b-a406-44d7-aeb8-09f440802b3a", linkInstance.ID)
	assert.Equal(t, "bbea65fd-b51f-464a-a5d3-dc228ff408c1", linkInstance.SchemaRelationshipID)
	linkElement, linkElementExists := reader.Schema.LinkedPropertyByName("address")
	assert.True(t, linkElementExists)
	assert.Equal(t, linkElement.ID, linkInstance.SchemaRelationshipID)

	assert.Equal(t, "7681b4f8-7d10-4855-8c87-7fef3b408c0b", linkInstance.From)
	assert.Equal(t, "e79e8d65-b094-4f36-94f2-1553cd84b4a2", linkInstance.To)
}
