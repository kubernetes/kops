package aws

import "github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"

type Tag struct {
	Key   *string `json:"tagKey,omitempty"`
	Value *string `json:"tagValue,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Tag) MarshalJSON() ([]byte, error) {
	type noMethod Tag
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Tag) SetKey(v *string) *Tag {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Tag) SetValue(v *string) *Tag {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}
