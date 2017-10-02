# Adding a validation to a resources schema

**Note:** By default, when creating a resource with `apiserver-boot create group version resource` a validation
function will be created in the versioned `<kind>_types.go` file.

To add server side validation for your resource override the `Validate` function
on the `<Kind>Strategy` struct in the version package.

Example:

File: `pkg/apis/<group>/<version>/bar_types.go`

```go
// Resource Validation
func (BarStrategy) Validate(ctx request.Context, obj runtime.Object) field.ErrorList {
	bar := obj.(*Bar)
	errors := field.ErrorList{}
	if ... {
		errors = append(errors, field.Invalid(
			field.NewPath("spec", "Field"),
			*bar.Spec.Field,
			"Error message"))
	}
	return errors
}
```

## Anatomy of validation

A default `<Kind>Strategy` is generated for each resource with an embedded
default Validation function.  To specify custom validation logic,
override the embedded implementation.

Cast the object type to your resource Kind

```go
bar := obj.(*Bar)
```

---

Use the field.Invalid function to specify errors scoped to fields in the object.

```go
field.Invalid(field.NewPath("spec", "Field"), *bar.Spec.Field, "Error message")
```

**Note:** To specify a different struct type for validation, specify `stragegy` in the resource
comment.  e.g. `// +resource:path=<resource>,strategy=<Kind>Strategy`.  This struct type must
have a single field of type `builders.DefaultStorageStrategy` for the generated code to correctly
create an pass it into the wiring.
