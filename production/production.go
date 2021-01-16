package production

import (
	"errors"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/uuid"
)

// Namespace is the canonical name of this extension.
const Namespace = "http://schemas.microsoft.com/3dmanufacturing/production/2015/06"

const (
	attrProdUUID = "UUID"
	attrPath     = "path"
)

var (
	ErrUUID             = errors.New("UUID MUST be any of the four UUID variants described in IETF RFC 4122")
	ErrProdRefInNonRoot = errors.New("non-root model file components MUST only reference objects in the same model file")
)

type Spec struct {
	LocalName string
}

func (e Spec) Namespace() string   { return Namespace }
func (e Spec) Required() bool      { return true }
func (e *Spec) SetRequired(r bool) {}
func (e *Spec) SetLocal(l string)  { e.LocalName = l }

func (e Spec) Local() string {
	if e.LocalName != "" {
		return e.LocalName
	}
	return "p"
}

// BuildAttr provides a UUID in the root model file build element to ensure
// that a 3MF package can be tracked across uses by various consumers.
type BuildAttr struct {
	UUID string
}

func GetBuildAttr(build *go3mf.Build) *BuildAttr {
	for _, a := range build.AnyAttr {
		if a, ok := a.(*BuildAttr); ok {
			return a
		}
	}
	return nil
}

// ObjectAttr provides a UUID in the item element
// for traceability across 3MF packages.
type ObjectAttr struct {
	UUID string
}

func GetObjectAttr(obj *go3mf.Object) *ObjectAttr {
	for _, a := range obj.AnyAttr {
		if a, ok := a.(*ObjectAttr); ok {
			return a
		}
	}
	return nil
}

// ItemAttr provides a UUID in the item element to ensure
// that each object can be reliably tracked.
type ItemAttr struct {
	UUID string
	Path string
}

func GetItemAttr(item *go3mf.Item) *ItemAttr {
	for _, a := range item.AnyAttr {
		if a, ok := a.(*ItemAttr); ok {
			return a
		}
	}
	return nil
}

// ObjectPath returns the Path extension attribute.
func (p *ItemAttr) ObjectPath() string {
	return p.Path
}

func (p *ItemAttr) getUUID() string {
	return p.UUID
}

// ObjectAttr provides a UUID in the component element
// for traceability across 3MF packages.
type ComponentAttr struct {
	UUID string
	Path string
}

func GetComponentAttr(comp *go3mf.Component) *ComponentAttr {
	for _, a := range comp.AnyAttr {
		if a, ok := a.(*ComponentAttr); ok {
			return a
		}
	}
	return nil
}

// ObjectPath returns the Path extension attribute.
func (p *ComponentAttr) ObjectPath() string {
	return p.Path
}

func (p *ComponentAttr) getUUID() string {
	return p.UUID
}

func AddMissingUUIDs(m *go3mf.Model) {
	if GetBuildAttr(&m.Build) == nil {
		m.Build.AnyAttr = append(m.Build.AnyAttr, &BuildAttr{UUID: uuid.New()})
	}
	for _, item := range m.Build.Items {
		ext := GetItemAttr(item)
		if ext == nil {
			item.AnyAttr = append(item.AnyAttr, &ItemAttr{
				UUID: uuid.New(),
			})
		} else if ext.UUID == "" {
			ext.UUID = uuid.New()
		}
	}
	m.WalkObjects(func(s string, obj *go3mf.Object) error {
		oext := GetObjectAttr(obj)
		if oext == nil {
			obj.AnyAttr = append(obj.AnyAttr, &ObjectAttr{UUID: uuid.New()})
		} else if oext.UUID == "" {
			oext.UUID = uuid.New()
		}
		for _, c := range obj.Components {
			ext := GetComponentAttr(c)
			if ext == nil {
				c.AnyAttr = append(c.AnyAttr, &ComponentAttr{
					UUID: uuid.New(),
				})
			} else if ext.UUID == "" {
				ext.UUID = uuid.New()
			}
		}
		return nil
	})
	return
}
