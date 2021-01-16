package production

import (
	"encoding/xml"

	"github.com/qmuntal/go3mf/spec/encoding"
)

// Marshal3MFAttr encodes the resource attributes.
func (u *BuildAttr) Marshal3MFAttr(_ encoding.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: u.UUID},
	}, nil
}

// Marshal3MFAttr encodes the resource attributes.
func (u *ObjectAttr) Marshal3MFAttr(_ encoding.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: u.UUID},
	}, nil
}

// Marshal3MFAttr encodes the resource attributes.
func (p *ItemAttr) Marshal3MFAttr(_ encoding.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrPath}, Value: p.Path},
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
	}, nil
}

// Marshal3MFAttr encodes the resource attributes.
func (p *ComponentAttr) Marshal3MFAttr(_ encoding.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrPath}, Value: p.Path},
		{Name: xml.Name{Space: Namespace, Local: attrProdUUID}, Value: p.UUID},
	}, nil
}
