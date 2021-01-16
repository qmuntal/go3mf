package production

import (
	"github.com/qmuntal/go3mf"
	specerr "github.com/qmuntal/go3mf/errors"
	"github.com/qmuntal/go3mf/spec/encoding"
	"github.com/qmuntal/go3mf/uuid"
)

func (e Spec) NewElementDecoder(_ encoding.ElementDecoderContext) encoding.ElementDecoder {
	return nil
}

func (e Spec) DecodeAttribute(parentNode interface{}, attr encoding.XMLAttr) (errs error) {
	switch t := parentNode.(type) {
	case *go3mf.Build:
		if attr.Name.Local == attrProdUUID {
			if err := uuid.Validate(string(attr.Value)); err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(attr.Name.Local, true))
			}
			t.AnyAttr = append(t.AnyAttr, &BuildAttr{UUID: string(attr.Value)})
		}
	case *go3mf.Item:
		switch attr.Name.Local {
		case attrProdUUID:
			if err := uuid.Validate(string(attr.Value)); err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(attr.Name.Local, true))
			}
			if ext := GetItemAttr(t); ext != nil {
				ext.UUID = string(attr.Value)
			} else {
				t.AnyAttr = append(t.AnyAttr, &ItemAttr{UUID: string(attr.Value)})
			}
		case attrPath:
			if ext := GetItemAttr(t); ext != nil {
				ext.Path = string(attr.Value)
			} else {
				t.AnyAttr = append(t.AnyAttr, &ItemAttr{Path: string(attr.Value)})
			}
		}
	case *go3mf.Object:
		if attr.Name.Local == attrProdUUID {
			if err := uuid.Validate(string(attr.Value)); err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(attr.Name.Local, true))
			}
			t.AnyAttr = append(t.AnyAttr, &ObjectAttr{UUID: string(attr.Value)})
		}
	case *go3mf.Component:
		switch attr.Name.Local {
		case attrProdUUID:
			if err := uuid.Validate(string(attr.Value)); err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(attr.Name.Local, true))
			}
			if ext := GetComponentAttr(t); ext != nil {
				ext.UUID = string(attr.Value)
			} else {
				t.AnyAttr = append(t.AnyAttr, &ComponentAttr{UUID: string(attr.Value)})
			}
		case attrPath:
			if ext := GetComponentAttr(t); ext != nil {
				ext.Path = string(attr.Value)
			} else {
				t.AnyAttr = append(t.AnyAttr, &ComponentAttr{Path: string(attr.Value)})
			}
		}
	}
	return
}
