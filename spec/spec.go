package spec

type PropertyGroup interface {
	Len() int
}

// Validator is the interface implemented by specs
// that can validate an element.
//
// Currently element can be a Model, Asset or Object.
type ValidatorSpec interface {
	Validate(path string, element interface{}) error
}
