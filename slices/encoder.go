package slices

import (
	"encoding/xml"
	"strconv"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/spec/encoding"
)

// Marshal3MFAttr encodes the resource attributes.
func (s *ObjectAttr) Marshal3MFAttr(_ encoding.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: Namespace, Local: attrSliceRefID}, Value: strconv.FormatUint(uint64(s.SliceStackID), 10)},
		{Name: xml.Name{Space: Namespace, Local: attrMeshRes}, Value: s.MeshResolution.String()},
	}, nil
}

// Marshal3MF encodes the resource.
func (s *SliceStack) Marshal3MF(x encoding.XMLEncoder) error {
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrSliceStack}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(s.ID), 10)},
	}}
	if s.BottomZ != 0 {
		xs.Attr = append(xs.Attr, xml.Attr{
			Name:  xml.Name{Local: attrZBottom},
			Value: strconv.FormatFloat(float64(s.BottomZ), 'f', x.FloatPresicion(), 32),
		})
	}
	x.EncodeToken(xs)
	for _, sl := range s.Slices {
		sl.marshal3MF(x)
	}
	x.SetAutoClose(true)
	for _, r := range s.Refs {
		x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrSliceRef}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrSliceRefID}, Value: strconv.FormatUint(uint64(r.SliceStackID), 10)},
			{Name: xml.Name{Local: attrSlicePath}, Value: r.Path},
		}})
	}
	x.SetAutoClose(false)
	x.EncodeToken(xs.End())
	return nil
}

func (s *Slice) marshal3MF(x encoding.XMLEncoder) {
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrSlice}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrZTop}, Value: strconv.FormatFloat(float64(s.TopZ), 'f', x.FloatPresicion(), 32)},
	}}
	x.EncodeToken(xs)

	marshalVertices(x, s.Vertices)
	marshalPolygons(x, s.Polygons)

	x.EncodeToken(xs.End())
}

func marshalPolygons(x encoding.XMLEncoder, ply []Polygon) {
	for _, p := range ply {
		xp := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrPolygon}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrStartV}, Value: strconv.FormatUint(uint64(p.StartV), 10)},
		}}
		x.EncodeToken(xp)
		x.SetAutoClose(true)
		for _, s := range p.Segments {
			xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrSegment}, Attr: []xml.Attr{
				{Name: xml.Name{Local: attrV2}, Value: strconv.FormatUint(uint64(s.V2), 10)},
			}}
			if s.PID != 0 {
				if s.P1 != s.P2 {
					xs.Attr = append(xs.Attr,
						xml.Attr{Name: xml.Name{Local: attrPID}, Value: strconv.FormatUint(uint64(s.PID), 10)},
						xml.Attr{Name: xml.Name{Local: attrP1}, Value: strconv.FormatUint(uint64(s.P1), 10)},
						xml.Attr{Name: xml.Name{Local: attrP2}, Value: strconv.FormatUint(uint64(s.P2), 10)},
					)
				} else {
					xs.Attr = append(xs.Attr,
						xml.Attr{Name: xml.Name{Local: attrPID}, Value: strconv.FormatUint(uint64(s.PID), 10)},
						xml.Attr{Name: xml.Name{Local: attrP1}, Value: strconv.FormatUint(uint64(s.P1), 10)},
					)
				}
			}
			x.EncodeToken(xs)
		}
		x.SetAutoClose(false)
		x.EncodeToken(xp.End())
	}
}

func marshalVertices(x encoding.XMLEncoder, vs []go3mf.Point2D) {
	xv := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrVertices}}
	x.EncodeToken(xv)
	x.SetAutoClose(true)
	prec := x.FloatPresicion()
	for _, v := range vs {
		x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrVertex}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrX}, Value: strconv.FormatFloat(float64(v.X()), 'f', prec, 32)},
			{Name: xml.Name{Local: attrY}, Value: strconv.FormatFloat(float64(v.Y()), 'f', prec, 32)},
		}})
	}
	x.SetAutoClose(false)
	x.EncodeToken(xv.End())
}
