package terraform

import "encoding/json"

type Literal struct {
	value string
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.value)
}

func LiteralExpression(s string) *Literal {
	return &Literal{value: s}
}

func LiteralSelfLink(resourceType, resourceName string) *Literal {
	return LiteralProperty(resourceType, resourceName, "self_link")
}

func LiteralProperty(resourceType, resourceName, prop string) *Literal {
	tfName := tfSanitize(resourceName)

	expr := "${" + resourceType + "." + tfName + "." + prop + "}"
	return LiteralExpression(expr)
}

func LiteralFromStringValue(s string) *Literal {
	return &Literal{value: s}
}
