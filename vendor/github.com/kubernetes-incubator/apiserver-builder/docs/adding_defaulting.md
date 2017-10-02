# Adding a value defaulting to a resources schema

To add server side field value defaulting for your resource override
the function `func (<group>.<Kind>SchemeFns) DefaultingFunction(o interface{})`
in the group package.

**Important:** The validation logic lives in the version package *not* the group package.

Example:

File: `pkg/apis/<group>/<version>/types_bar.go`

```go
func (BarSchemeFns) DefaultingFunction(o interface{}) {
	obj := o.(*Bar)
	if obj.Spec.Field == nil {
		f := "value"
		obj.Spec.Field = &f
	}
}
```

## Anatomy of defaulting

A default `<group>.<Kind>SchemeFns` is generated for each resource with an embedded
empty defaulting function.  To specify custom defaulting logic,
override the embedded implementation.

Cast the object type to your resource Kind

```go
bar := obj.(*Bar)
```

---

Update set values for fields with nil values.

```go
	if obj.Spec.Field == nil {
		f := "value"
		obj.Spec.Field = &f
	}
```