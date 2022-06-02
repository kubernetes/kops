package azure

import "github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"

// region Tag

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

// endregion

// region OSDisk

type OSDisk struct {
	SizeGB                  *int    `json:"sizeGB,omitempty"`
	Type                    *string `json:"type,omitempty"`
	UtilizeEphemeralStorage *bool   `json:"utilizeEphemeralStorage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o OSDisk) MarshalJSON() ([]byte, error) {
	type noMethod OSDisk
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *OSDisk) SetSizeGB(v *int) *OSDisk {
	if o.SizeGB = v; o.SizeGB == nil {
		o.nullFields = append(o.nullFields, "SizeGB")
	}
	return o
}

func (o *OSDisk) SetType(v *string) *OSDisk {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *OSDisk) SetUtilizeEphemeralStorage(v *bool) *OSDisk {
	if o.UtilizeEphemeralStorage = v; o.UtilizeEphemeralStorage == nil {
		o.nullFields = append(o.nullFields, "UtilizeEphemeralStorage")
	}
	return o
}

// endregion

// region Label

type Label struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Label) MarshalJSON() ([]byte, error) {
	type noMethod Label
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Label) SetKey(v *string) *Label {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Label) SetValue(v *string) *Label {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region Taint

type Taint struct {
	Key    *string `json:"key,omitempty"`
	Value  *string `json:"value,omitempty"`
	Effect *string `json:"effect,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Taint) MarshalJSON() ([]byte, error) {
	type noMethod Taint
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Taint) SetKey(v *string) *Taint {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Taint) SetValue(v *string) *Taint {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

func (o *Taint) SetEffect(v *string) *Taint {
	if o.Effect = v; o.Effect == nil {
		o.nullFields = append(o.nullFields, "Effect")
	}
	return o
}

// endregion
