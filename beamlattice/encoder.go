package beamlattice

import (
	"encoding/xml"
	"strconv"

	"github.com/qmuntal/go3mf/spec/encoding"
)

// Marshal3MF encodes the resource.
func (m *BeamLattice) Marshal3MF(x encoding.XMLEncoder) error {
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrBeamLattice}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrMinLength}, Value: strconv.FormatFloat(float64(m.MinLength), 'f', x.FloatPresicion(), 32)},
		{Name: xml.Name{Local: attrRadius}, Value: strconv.FormatFloat(float64(m.Radius), 'f', x.FloatPresicion(), 32)},
	}}
	if m.ClipMode != ClipNone {
		xs.Attr = append(xs.Attr, xml.Attr{Name: xml.Name{Local: attrClippingMode}, Value: m.ClipMode.String()})
	}
	if m.ClippingMeshID != 0 {
		xs.Attr = append(xs.Attr, xml.Attr{
			Name:  xml.Name{Local: attrClippingMesh},
			Value: strconv.FormatUint(uint64(m.ClippingMeshID), 10)},
		)
	}
	if m.RepresentationMeshID != 0 {
		xs.Attr = append(xs.Attr, xml.Attr{
			Name:  xml.Name{Local: attrRepresentationMesh},
			Value: strconv.FormatUint(uint64(m.RepresentationMeshID), 10)},
		)
	}
	if m.CapMode != CapModeSphere {
		xs.Attr = append(xs.Attr, xml.Attr{Name: xml.Name{Local: attrCap}, Value: m.CapMode.String()})
	}
	x.EncodeToken(xs)

	marshalBeams(x, m)
	marshalBeamsets(x, m)

	x.EncodeToken(xs.End())
	return nil
}

func marshalBeamsets(x encoding.XMLEncoder, m *BeamLattice) {
	xb := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrBeamSets}}
	x.EncodeToken(xb)
	for _, bs := range m.BeamSets {
		xbs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrBeamSet}}
		if bs.Name != "" {
			xbs.Attr = append(xbs.Attr, xml.Attr{Name: xml.Name{Local: attrName}, Value: bs.Name})
		}
		if bs.Identifier != "" {
			xbs.Attr = append(xbs.Attr, xml.Attr{Name: xml.Name{Local: attrIdentifier}, Value: bs.Identifier})
		}
		x.EncodeToken(xbs)
		x.SetAutoClose(true)
		for _, ref := range bs.Refs {
			x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrRef}, Attr: []xml.Attr{
				{Name: xml.Name{Local: attrIndex}, Value: strconv.FormatUint(uint64(ref), 10)},
			}})
		}
		x.SetAutoClose(false)
		x.EncodeToken(xbs.End())
	}
	x.EncodeToken(xb.End())
}

func marshalBeams(x encoding.XMLEncoder, m *BeamLattice) {
	xb := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrBeams}}
	x.EncodeToken(xb)
	x.SetAutoClose(true)
	for _, b := range m.Beams {
		xbeam := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrBeam}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrV1}, Value: strconv.FormatUint(uint64(b.Indices[0]), 10)},
			{Name: xml.Name{Local: attrV2}, Value: strconv.FormatUint(uint64(b.Indices[1]), 10)},
		}}
		if b.Radius[0] > 0 && b.Radius[0] != m.Radius {
			xbeam.Attr = append(xbeam.Attr, xml.Attr{
				Name:  xml.Name{Local: attrR1},
				Value: strconv.FormatFloat(float64(b.Radius[0]), 'f', x.FloatPresicion(), 32),
			})
		}
		if b.Radius[1] > 0 && b.Radius[1] != m.Radius {
			xbeam.Attr = append(xbeam.Attr, xml.Attr{
				Name:  xml.Name{Local: attrR2},
				Value: strconv.FormatFloat(float64(b.Radius[1]), 'f', x.FloatPresicion(), 32),
			})
		}
		if b.CapMode[0] != m.CapMode {
			xbeam.Attr = append(xbeam.Attr, xml.Attr{Name: xml.Name{Local: attrCap1}, Value: b.CapMode[0].String()})
		}
		if b.CapMode[1] != m.CapMode {
			xbeam.Attr = append(xbeam.Attr, xml.Attr{Name: xml.Name{Local: attrCap2}, Value: b.CapMode[1].String()})
		}
		x.EncodeToken(xbeam)
	}
	x.SetAutoClose(false)
	x.EncodeToken(xb.End())
}
