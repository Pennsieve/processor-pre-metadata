package paths

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPropertiesFilePath(t *testing.T) {
	modelID := uuid.NewString()
	assert.Equal(t, fmt.Sprintf("%s/%s/%s.json", SchemaDirectory, PropertiesDirectory, modelID), PropertiesFilePath(modelID))
}
