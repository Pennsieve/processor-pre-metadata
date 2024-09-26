package client

import (
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
	assert.Len(t, records, 2)
}
