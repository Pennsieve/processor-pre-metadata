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

	propNameToProp := idToPropNameToProp["5b07e038-9829-46c9-b698-bf4efef81341"]
	assert.Len(t, propNameToProp, 7)

	// Name
	name := propNameToProp["name"]

	nameDataType, err := name.DecodeDataType()
	require.NoError(t, err)
	assert.Equal(t, instance.StringType, nameDataType)

	nameValue, err := name.DecodeValue()
	require.NoError(t, err)
	assert.Equal(t, "stone", nameValue)

	// ID
	id := propNameToProp["id"]

	idDataType, err := id.DecodeDataType()
	require.NoError(t, err)
	assert.Equal(t, instance.LongType, idDataType)

	idValue, err := id.DecodeValue()
	require.NoError(t, err)
	assert.Equal(t, int64(1), idValue)

}
