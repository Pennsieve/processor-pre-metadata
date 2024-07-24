package schema

import (
	"log/slog"
)

type Model struct {
	Element
	Properties       []Property
	LinkedProperties []LinkedProperty
	Relationships    []Relationship
}

func ModelFromMap(jsonMap map[string]any) (*Model, error) {
	if isModel, err := IsModel(jsonMap); err != nil {
		return nil, err
	} else if !isModel {
		return nil, nil
	}
	return &Model{
		Element: ElementFromMap(jsonMap),
	}, nil
}

func (m *Model) Logger(logger *slog.Logger) *slog.Logger {
	return logger.With(slog.Group("model",
		slog.String("id", m.ID),
		slog.String("name", m.Name)))
}
