package spec

type PropertyGroup interface {
	Len() int
}

// Validator is the interface implemented by specs
// that can validate an element.
//
// model is guaranteed to be a *Model.
// element can be a Model, Asset or Object.
// In the future this list can be expanded.
type ValidatorSpec interface {
	Validate(model interface{}, path string, element interface{}) error
}
