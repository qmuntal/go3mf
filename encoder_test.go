package go3mf

import (
	"bytes"
	"encoding/xml"
	"errors"
	"image/color"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-test/deep"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
	"github.com/stretchr/testify/mock"
)

func (f *fakeAttr) Marshal3MFAttr(_ encoding.XMLEncoder) ([]xml.Attr, error) {
	return []xml.Attr{
		{Name: xml.Name{Space: fakeExtension, Local: "value"}, Value: f.Value},
	}, nil
}

// Marshal3MF encodes the resource.
func (f *fakeAsset) Marshal3MF(x encoding.XMLEncoder) error {
	xs := xml.StartElement{Name: xml.Name{Space: fakeExtension, Local: "fakeasset"}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(f.ID), 10)},
	}}
	x.EncodeToken(xs)
	x.EncodeToken(xs.End())
	return nil
}

type mockPackagePart struct {
	mock.Mock
}

func (m *mockPackagePart) Write(arg1 []byte) (int, error) {
	args := m.Called(arg1)
	return args.Int(0), args.Error(1)
}

func (m *mockPackagePart) AddRelationship(args0 Relationship) {
	m.Called(args0)
}

func TestMarshalModel(t *testing.T) {
	m := &Model{
		Units: UnitMillimeter, Language: "en-US", Path: "/3D/3dmodel.model", Thumbnail: "/thumbnail.png",
		Specs:   map[string]spec.Spec{fakeExtension: &fakeSpec{}},
		AnyAttr: AnyAttr{&fakeAttr{Value: "model_fake"}},
		Resources: Resources{
			Assets: []Asset{
				&BaseMaterials{ID: 5, Materials: []Base{
					{Name: "Blue PLA", Color: color.RGBA{0, 0, 255, 255}},
					{Name: "Red ABS", Color: color.RGBA{255, 0, 0, 255}},
				}}, &fakeAsset{ID: 25}},
			Objects: []*Object{
				{
					ID: 8, Name: "Box 1", PartNumber: "11111111-1111-1111-1111-111111111111", Thumbnail: "/a.png",
					AnyAttr: AnyAttr{&fakeAttr{Value: "object_fake"}},
					PID:     1, PIndex: 1, Type: ObjectTypeModel, Mesh: &Mesh{
						Vertices: []Point3D{
							{0, 0, 0}, {100, 0, 0}, {100, 100, 0},
							{0, 100, 0}, {0, 0, 100}, {100, 0, 100},
							{100, 100, 100}, {0, 100, 100}},
						Triangles: []Triangle{
							NewTrianglePID(3, 2, 1, 5, 0, 0, 0),
							NewTrianglePID(1, 0, 3, 5, 0, 0, 0),
							NewTrianglePID(4, 5, 6, 5, 1, 1, 1),
							NewTrianglePID(6, 7, 4, 5, 1, 1, 1),
							NewTrianglePID(0, 1, 5, 5, 0, 1, 2),
							NewTrianglePID(5, 4, 0, 5, 3, 0, 2),
							NewTrianglePID(1, 2, 6, 5, 0, 1, 2),
							NewTrianglePID(6, 5, 1, 5, 2, 1, 3),
							NewTrianglePID(2, 3, 7, 5, 0, 0, 0),
							NewTrianglePID(7, 6, 2, 5, 0, 0, 0),
							NewTrianglePID(3, 0, 4, 5, 0, 0, 0),
							NewTrianglePID(4, 7, 3, 5, 0, 0, 0),
						},
					}},
				{
					ID: 20, Type: ObjectTypeSupport,
					Metadata: []Metadata{{Name: xml.Name{Space: "qm", Local: "CustomMetadata3"}, Type: "xs:boolean", Value: "1"}, {Name: xml.Name{Space: "qm", Local: "CustomMetadata4"}, Type: "xs:boolean", Value: "2"}},
					Components: []*Component{{ObjectID: 8, Transform: Matrix{3, 0, 0, 0, 0, 1, 0, 0, 0, 0, 2, 0, -66.4, -87.1, 8.8, 1},
						AnyAttr: AnyAttr{&fakeAttr{Value: "component_fake"}}}},
				},
			},
		},
		Build: Build{
			AnyAttr: AnyAttr{&fakeAttr{Value: "build_fake"}},
			Items: []*Item{
				{
					ObjectID: 20, PartNumber: "bob", Transform: Matrix{1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, -66.4, -87.1, 8.8, 1},
					Metadata: []Metadata{{Name: xml.Name{Space: "qm", Local: "CustomMetadata3"}, Type: "xs:boolean", Value: "1"}},
				},
				{ObjectID: 21, AnyAttr: AnyAttr{&fakeAttr{Value: "item_fake"}}},
			}}, Metadata: []Metadata{
			{Name: xml.Name{Local: "Application"}, Value: "go3mf app"},
			{Name: xml.Name{Space: "qm", Local: "CustomMetadata1"}, Preserve: true, Type: "xs:string", Value: "CE8A91FB-C44E-4F00-B634-BAA411465F6A"},
		}}

	t.Run("base", func(t *testing.T) {
		b, err := MarshalModel(m)
		if err != nil {
			t.Errorf("MarshalModel() error = %v", err)
			return
		}
		newModel := new(Model)
		newModel.WithSpec(&fakeSpec{})
		newModel.Path = m.Path
		if err := UnmarshalModel(b, newModel); err != nil {
			t.Errorf("MarshalModel() error decoding = %v, s = %s", err, string(b))
			return
		}
		if diff := deep.Equal(m, newModel); diff != nil {
			t.Errorf("MarshalModel() = %v, s = %s", diff, string(b))
		}
	})
}

func TestEncoder_writeAttachements(t *testing.T) {
	type args struct {
		m *Model
	}
	tests := []struct {
		name    string
		e       *Encoder
		args    args
		wantErr bool
	}{
		{"empty", &Encoder{}, args{new(Model)}, false},
		{"err-create", &Encoder{}, args{&Model{Attachments: []Attachment{{}}}}, true},
		{"base", &Encoder{}, args{&Model{Attachments: []Attachment{{Stream: new(bytes.Buffer)}}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var argErr error
			if tt.wantErr {
				argErr = errors.New("")
			}
			m := new(mockPackage)
			mp := new(mockPackagePart)
			mp.On("Write", mock.Anything).Return(mock.Anything, mock.Anything)
			m.On("Create", mock.Anything, mock.Anything).Return(mp, argErr)
			m.On("AddRelationship", mock.Anything).Return()
			tt.e.w = m
			if err := tt.e.writeAttachements(tt.args.m.Attachments); (err != nil) != tt.wantErr {
				t.Errorf("Encoder.writeAttachements() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEncoder(t *testing.T) {
	tests := []struct {
		name string
		want *Encoder
	}{
		{"base", &Encoder{FloatPrecision: defaultFloatPrecision, w: newOpcWriter(nil)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEncoder(nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEncoder() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestEncoder_Encode_Normalize(t *testing.T) {
	type args struct {
		m *Model
	}
	tests := []struct {
		name string
		args args
		want *Model
	}{
		{"empty", args{new(Model)}, &Model{Path: DefaultModelPath}},
		{"withAttrs", args{&Model{Path: "a/other.ml", Thumbnail: "/Metadata/thumbnail.png", Attachments: []Attachment{
			{ContentType: "image/png", Path: "Metadata/thumbnail.png", Stream: bytes.NewBufferString("fake")},
		}}}, &Model{Path: "/a/other.ml", Units: UnitMillimeter, Thumbnail: "/Metadata/thumbnail.png", Attachments: []Attachment{
			{ContentType: "image/png", Path: "/Metadata/thumbnail.png", Stream: bytes.NewBufferString("fake")},
		}, RootRelationships: []Relationship{
			{Path: "/Metadata/thumbnail.png", Type: RelTypeThumbnail, ID: "rId1"},
		}}},
		{"withRootRel", args{&Model{
			RootRelationships: []Relationship{
				{Path: "Metadata/thumbnail.png", Type: RelTypeThumbnail, ID: "2"},
				{Path: "Metadata/thumbnail.png", Type: RelTypeThumbnail, ID: "2"},
			},
			Attachments: []Attachment{
				{ContentType: "application/vnd.ms-printing.printticket+xml", Path: "/3D/Metadata/pt.xml", Stream: bytes.NewBufferString("other")},
				{ContentType: "image/png", Path: "/Metadata/thumbnail.png", Stream: bytes.NewBufferString("fake")},
			}}},
			&Model{Path: DefaultModelPath,
				RootRelationships: []Relationship{
					{Path: "Metadata/thumbnail.png", Type: RelTypeThumbnail, ID: "2"},
				},
				Attachments: []Attachment{
					{ContentType: "image/png", Path: "/Metadata/thumbnail.png", Stream: bytes.NewBufferString("fake")},
				}}},
		{"withChildModel", args{&Model{
			Childs: map[string]*ChildModel{
				"empty.model": {},
				"/other.model": {Relationships: []Relationship{
					{Path: "/3D/Metadata/pt.xml", Type: "http://schemas.microsoft.com/3dmanufacturing/2013/01/printticket", ID: "1"}},
				},
			}}}, &Model{Path: DefaultModelPath,
			Childs: map[string]*ChildModel{
				"/3D/empty.model": {},
				"/other.model":    {},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := new(bytes.Buffer)
			if err := NewEncoder(buff).Encode(tt.args.m); err != nil {
				t.Errorf("Encoder.Encode() error = %v", err)
				return
			}
			newModel := new(Model)
			err := NewDecoder(bytes.NewReader(buff.Bytes()), int64(buff.Len())).Decode(newModel)
			if err != nil {
				t.Errorf("Encoder.Encode() malformed = %v", err)
				return
			}
			if tt.args.m.Path == "" {
				tt.args.m.Path = DefaultModelPath
			}
			if diff := deep.Equal(newModel, tt.want); diff != nil {
				t.Errorf("MarshalModel() = %v", diff)
			}
		})
	}
}

func TestEncoder_Encode_Roundtrip(t *testing.T) {
	type args struct {
		m *Model
	}
	tests := []struct {
		name string
		args args
	}{
		{"empty", args{new(Model)}},
		{"withAttrs", args{&Model{Path: "/a/other.ml", Language: "un", Units: UnitFoot, Thumbnail: "/thumb.png"}}},
		{"withMetdata", args{&Model{Metadata: []Metadata{
			{Name: xml.Name{Local: "a"}, Value: "b", Type: "tp", Preserve: true},
			{Name: xml.Name{Local: "ab"}, Value: "bb", Type: "tpb", Preserve: false},
		}}}},
		{"withRootRel", args{&Model{
			RootRelationships: []Relationship{
				{Path: "/3D/Metadata/pt.xml", Type: "http://schemas.microsoft.com/3dmanufacturing/2013/01/printticket", ID: "1"},
				{Path: "/Metadata/thumbnail.png", Type: "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail", ID: "2"},
			},
			Attachments: []Attachment{
				{ContentType: "application/vnd.ms-printing.printticket+xml", Path: "/3D/Metadata/pt.xml", Stream: bytes.NewBufferString("other")},
				{ContentType: "image/png", Path: "/Metadata/thumbnail.png", Stream: bytes.NewBufferString("fake")},
			}}},
		},
		{"withChildModel", args{&Model{
			Attachments: []Attachment{
				{ContentType: "application/vnd.ms-printing.printticket+xml", Path: "/3D/Metadata/pt.xml", Stream: bytes.NewBufferString("other")},
			},
			Childs: map[string]*ChildModel{
				"/empty.model": {},
				"/other.model": {Relationships: []Relationship{
					{Path: "/3D/Metadata/pt.xml", Type: "http://schemas.microsoft.com/3dmanufacturing/2013/01/printticket", ID: "1"}},
				},
			}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := new(bytes.Buffer)
			if err := NewEncoder(buff).Encode(tt.args.m); err != nil {
				t.Errorf("Encoder.Encode() error = %v", err)
				return
			}
			newModel := new(Model)
			err := NewDecoder(bytes.NewReader(buff.Bytes()), int64(buff.Len())).Decode(newModel)
			if err != nil {
				t.Errorf("Encoder.Encode() malformed = %v", err)
				return
			}
			if tt.args.m.Path == "" {
				tt.args.m.Path = DefaultModelPath
			}
			if diff := deep.Equal(newModel, tt.args.m); diff != nil {
				t.Errorf("MarshalModel() = %v", diff)
			}
		})
	}
}
