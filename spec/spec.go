package spec

type Spec interface {
	Namespace() string
	Local() string
	Required() bool
	SetRequired(bool)
	SetLocal(string)
}

type UnknownSpec struct {
	SpaceName  string
	LocalName  string
	IsRequired bool
}

func (u *UnknownSpec) Namespace() string  { return u.SpaceName }
func (u *UnknownSpec) Local() string      { return u.LocalName }
func (u *UnknownSpec) Required() bool     { return u.IsRequired }
func (u *UnknownSpec) SetLocal(l string)  { u.LocalName = l }
func (u *UnknownSpec) SetRequired(r bool) { u.IsRequired = r }

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
