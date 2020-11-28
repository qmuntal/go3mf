package production

import (
	"encoding/xml"

	"github.com/qmuntal/go3mf"
)

// Namespace is the canonical name of this extension.
const Namespace = "http://schemas.microsoft.com/3dmanufacturing/production/2015/06"

type Spec struct {
	LocalName  string
	IsRequired bool
}

func (e Spec) Namespace() string   { return Namespace }
func (e Spec) Required() bool      { return e.IsRequired }
func (e *Spec) SetRequired(r bool) { e.IsRequired = r }
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

// Marshal3MFAttr encodes the resource attributes.
func (u *BuildAttr) Marshal3MFAttr(_ *go3mf.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: u.UUID},
	}, nil
}

// ObjectAttr provides a UUID in the item element
// for traceability across 3MF packages.
type ObjectAttr struct {
	UUID string
}

// Marshal3MFAttr encodes the resource attributes.
func (u *ObjectAttr) Marshal3MFAttr(_ *go3mf.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: u.UUID},
	}, nil
}

// ItemAttr provides a UUID in the item element to ensure
// that each object can be reliably tracked.
type ItemAttr struct {
	UUID string
	Path string
}

// ObjectPath returns the Path extension attribute.
func (p *ItemAttr) ObjectPath() string {
	return p.Path
}

func (p *ItemAttr) getUUID() string {
	return p.UUID
}

// Marshal3MFAttr encodes the resource attributes.
func (p *ItemAttr) Marshal3MFAttr(_ *go3mf.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrPath}, Value: p.Path},
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
	}, nil
}

// ObjectAttr provides a UUID in the component element
// for traceability across 3MF packages.
type ComponentAttr struct {
	UUID string
	Path string
}

// ObjectPath returns the Path extension attribute.
func (p *ComponentAttr) ObjectPath() string {
	return p.Path
}

func (p *ComponentAttr) getUUID() string {
	return p.UUID
}

// Marshal3MFAttr encodes the resource attributes.
func (p *ComponentAttr) Marshal3MFAttr(_ *go3mf.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrPath}, Value: p.Path},
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
	}, nil
}

const (
	attrProdUUID = "UUID"
	attrPath     = "path"
)
