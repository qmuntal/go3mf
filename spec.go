package go3mf

type Spec interface {
	Namespace() string
	Local() string
	Required() bool
	SetRequired(bool)
	SetLocal(string)
	SetModel(*Model)
}

type UnknownSpec struct {
	SpaceName  string
	LocalName  string
	IsRequired bool
	m          *Model
}

func (u *UnknownSpec) Namespace() string  { return u.SpaceName }
func (u *UnknownSpec) Local() string      { return u.LocalName }
func (u *UnknownSpec) Required() bool     { return u.IsRequired }
func (u *UnknownSpec) SetLocal(l string)  { u.LocalName = l }
func (u *UnknownSpec) SetRequired(r bool) { u.IsRequired = r }
func (u *UnknownSpec) SetModel(m *Model)  { u.m = m }

type objectPather interface {
	ObjectPath() string
}

type propertyGroup interface {
	Len() int
}
