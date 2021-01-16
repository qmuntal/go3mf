package validation

import "github.com/qmuntal/go3mf/spec"

type PropertyGroup interface {
	Len() int
}

// ValidatorSpec is the interface implemented by specs
// that can validate an element.
//
// model is guaranteed to be a *Model.
// element can be a Model, Asset or Object.
// In the future this list can be expanded.
type ValidatorSpec interface {
	spec.Spec
	Validate(model interface{}, path string, element interface{}) error
}
