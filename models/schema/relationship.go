package schema

import "log/slog"

const FromKey = "from"
const ToKey = "to"
const PositionKey = "position"

type Relationship struct {
	Element
	From string `json:"from"`
	To   string `json:"to"`
}

func (m *Relationship) Logger(logger *slog.Logger) *slog.Logger {
	return logger.With(slog.Group("relationship",
		slog.String("id", m.ID),
		slog.String("displayName", m.DisplayName)))
}

func RelationshipFromMap(jsonMap map[string]any) (*Relationship, error) {
	if isRelationship, err := IsRelationship(jsonMap); err != nil {
		return nil, err
	} else if !isRelationship {
		return nil, nil
	}
	return &Relationship{
		Element: ElementFromMap(jsonMap),
		From:    jsonMap[FromKey].(string),
		To:      jsonMap[ToKey].(string),
	}, nil
}

type LinkedProperty struct {
	Element
	From     string `json:"from"`
	To       string `json:"to"`
	Position int    `json:"position"`
}

func (m *LinkedProperty) Logger(logger *slog.Logger) *slog.Logger {
	return logger.With(slog.Group("linkedProperties",
		slog.String("id", m.ID),
		slog.String("displayName", m.DisplayName)))
}

func LinkedPropertyFromMap(jsonMap map[string]any) (*LinkedProperty, error) {
	if isLinkedProp, err := IsLinkedProperty(jsonMap); err != nil {
		return nil, err
	} else if !isLinkedProp {
		return nil, nil
	}
	return &LinkedProperty{
		Element: ElementFromMap(jsonMap),
		From:    jsonMap[FromKey].(string),
		To:      jsonMap[ToKey].(string),
		// when unmarshalling to any, a JSON number will be treated as float64
		Position: int(jsonMap[PositionKey].(float64)),
	}, nil
}
