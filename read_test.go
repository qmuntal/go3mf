package go3mf

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-test/deep"
	specerr "github.com/qmuntal/go3mf/errors"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
	"github.com/qmuntal/go3mf/spec/validation"
	"github.com/stretchr/testify/mock"
)

const fakeExtension = "http://dummy.com/fake_ext"

var _ validation.ValidatorSpec = new(fakeSpec)
var _ encoding.DecoderSpec = new(fakeSpec)
var _ encoding.CharDataElementDecoder = new(metadataDecoder)
var _ encoding.ChildElementDecoder = new(baseMaterialsDecoder)

type fakeSpec struct {
	m *Model
}

func (f *fakeSpec) Namespace() string  { return fakeExtension }
func (f *fakeSpec) Required() bool     { return true }
func (f *fakeSpec) Local() string      { return "qm" }
func (f *fakeSpec) SetLocal(_ string)  {}
func (f *fakeSpec) SetRequired(_ bool) {}
func (f *fakeSpec) SetModel(m *Model)  { f.m = m }

func (f *fakeSpec) NewElementDecoder(ctx encoding.ElementDecoderContext) encoding.ElementDecoder {
	if e, ok := ctx.ParentElement.(*Resources); ok {
		return &fakeAssetDecoder{resources: e}
	}
	return nil
}

func (f *fakeSpec) DecodeAttribute(parentNode interface{}, attr encoding.XMLAttr) error {
	switch t := parentNode.(type) {
	case *Object:
		t.AnyAttr = append(t.AnyAttr, &fakeAttr{string(attr.Value)})
	case *Build:
		t.AnyAttr = append(t.AnyAttr, &fakeAttr{string(attr.Value)})
	case *Model:
		t.AnyAttr = append(t.AnyAttr, &fakeAttr{string(attr.Value)})
	case *Item:
		t.AnyAttr = append(t.AnyAttr, &fakeAttr{string(attr.Value)})
	case *Component:
		t.AnyAttr = append(t.AnyAttr, &fakeAttr{string(attr.Value)})
	}
	return nil
}

func (f *fakeSpec) Validate(_ interface{}, _ string, e interface{}) error {
	if e, ok := e.(*Model); ok {
		var errs []error
		if len(e.Build.AnyAttr) == 1 {
			if _, ok := e.Build.AnyAttr[0].(*fakeAttr); ok {
				errs = append(errs, errors.New("Build: fake"))
			}
		}
		return specerr.Append(nil, errs...)
	}
	return nil
}

type fakeAsset struct {
	ID uint32
}

func (f *fakeAsset) Identify() uint32 {
	return f.ID
}

type fakeAttr struct {
	Value string
}

func (f *fakeAttr) ObjectPath() string { return f.Value }

type fakeAssetDecoder struct {
	baseDecoder
	resources *Resources
}

func (f *fakeAssetDecoder) Start(att []encoding.XMLAttr) error {
	id, _ := strconv.ParseUint(string(att[0].Value), 10, 32)
	f.resources.Assets = append(f.resources.Assets, &fakeAsset{ID: uint32(id)})
	return nil
}

type modelBuilder struct {
	str      strings.Builder
	hasModel bool
}

func (m *modelBuilder) withElement(s string) *modelBuilder {
	m.str.WriteString(s)
	m.str.WriteString("\n")
	return m
}

func (m *modelBuilder) addAttr(prefix, name, value string) *modelBuilder {
	if prefix != "" {
		m.str.WriteString(fmt.Sprintf(`%s:`, prefix))
	}
	if name != "" {
		m.str.WriteString(fmt.Sprintf(`%s="%s" `, name, value))
	}
	return m
}

func (m *modelBuilder) withDefaultModel() *modelBuilder {
	m.withModel("millimeter", "en-US", "/thumbnail.png")
	return m
}

func (m *modelBuilder) withModel(unit string, lang string, thumbnail string) *modelBuilder {
	m.str.WriteString(`<model `)
	m.addAttr("", "unit", unit).addAttr("xml", "lang", lang)
	m.addAttr("", "xmlns", Namespace).addAttr("xmlns", "qm", fakeExtension)
	m.addAttr("", "requiredextensions", "qm")
	if thumbnail != "" {
		m.addAttr("", "thumbnail", thumbnail)
	}
	m.str.WriteString(">\n")
	m.hasModel = true
	return m
}

func (m *modelBuilder) withEncoding(encode string) *modelBuilder {
	m.str.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="%s"?>`, encode))
	m.str.WriteString("\n")
	return m
}

func (m *modelBuilder) build(name string) *mockFile {
	if name == "" {
		name = "/3D/3dmodel.model"
	}
	if m.hasModel {
		m.str.WriteString("</model>\n")
	}
	f := new(mockFile)
	f.On("Name").Return(name).Maybe()
	f.On("Open").Return(ioutil.NopCloser(bytes.NewBufferString(m.str.String())), nil).Maybe()
	return f
}

type mockFile struct {
	mock.Mock
}

func newMockFile(name string, relationships []Relationship, other *mockFile, openErr bool) *mockFile {
	m := new(mockFile)
	m.On("Name").Return(name).Maybe()
	m.On("ContentType").Return("").Maybe()
	m.On("Relationships").Return(relationships).Maybe()
	m.On("FindFileFromName", mock.Anything).Return(other, other != nil).Maybe()
	var err error
	if openErr {
		err = errors.New("")
	}
	m.On("Open").Return(ioutil.NopCloser(new(bytes.Buffer)), err).Maybe()
	return m
}

func (m *mockFile) Open() (io.ReadCloser, error) {
	args := m.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFile) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFile) ContentType() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockFile) FindFileFromName(args0 string) (packageFile, bool) {
	args := m.Called(args0)
	return args.Get(0).(packageFile), args.Bool(1)
}

func (m *mockFile) Relationships() []Relationship {
	args := m.Called()
	return args.Get(0).([]Relationship)
}

type mockPackage struct {
	mock.Mock
}

func newMockPackage(other *mockFile) *mockPackage {
	m := new(mockPackage)
	m.On("Open", mock.Anything).Return(nil).Maybe()
	m.On("Create", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	m.On("Relationships").Return([]Relationship{{Path: DefaultModelPath, Type: RelType3DModel}}).Maybe()
	m.On("FindFileFromName", mock.Anything).Return(other, other != nil).Maybe()
	return m
}

func (m *mockPackage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockPackage) Relationships() []Relationship {
	args := m.Called()
	return args.Get(0).([]Relationship)
}

func (m *mockPackage) AddRelationship(args0 Relationship) {
	m.Called(args0)
}

func (m *mockPackage) Create(args0, args1 string) (packagePart, error) {
	args := m.Called(args0, args1)
	return args.Get(0).(packagePart), args.Error(1)
}

func (m *mockPackage) Open(f func(r io.Reader) io.ReadCloser) error {
	args := m.Called(f)
	return args.Error(0)
}

func (m *mockPackage) FindFileFromName(args0 string) (packageFile, bool) {
	args := m.Called(args0)
	return args.Get(0).(packageFile), args.Bool(1)
}

func TestDecoder_processOPC(t *testing.T) {
	extType := "fake_type"
	otherModel := newMockFile("/other.model", nil, nil, false)
	tests := []struct {
		name    string
		d       *Decoder
		want    *Model
		wantErr bool
	}{
		{"noRoot", &Decoder{p: newMockPackage(nil)}, &Model{}, true},
		{"noRels", &Decoder{p: newMockPackage(newMockFile("/a.model", nil, nil, false))}, &Model{Path: "/a.model"}, false},
		{"withThumb", &Decoder{
			p: newMockPackage(newMockFile("/a.model", []Relationship{{Type: RelTypeThumbnail, Path: "/a.png"}}, newMockFile("/a.png", nil, nil, false), false)),
		}, &Model{
			Path:          "/a.model",
			Relationships: []Relationship{{Path: "/a.png", Type: RelTypeThumbnail}},
			Attachments:   []Attachment{{Path: "/a.png", Stream: new(bytes.Buffer)}},
		}, false},
		{"withPrintTicket", &Decoder{
			p: newMockPackage(newMockFile("/a.model", []Relationship{{Type: RelTypePrintTicket, Path: "/pc.png"}}, newMockFile("/pc.png", nil, nil, false), false)),
		}, &Model{
			Path:          "/a.model",
			Relationships: []Relationship{{Path: "/pc.png", Type: RelTypePrintTicket}},
			Attachments:   []Attachment{{Path: "/pc.png", Stream: new(bytes.Buffer)}},
		}, false},
		{"withExtRel", &Decoder{
			p: newMockPackage(newMockFile("/a.model", []Relationship{{Type: extType, Path: "/other.png"}}, newMockFile("/other.png", nil, nil, false), false)),
		}, &Model{
			Path:          "/a.model",
			Relationships: []Relationship{{Path: "/other.png", Type: extType}},
			Attachments:   []Attachment{{Path: "/other.png", Stream: new(bytes.Buffer)}},
		}, false},
		{"withOtherRel", &Decoder{
			p: newMockPackage(newMockFile("/a.model", []Relationship{{Type: "other", Path: "/a.png"}}, nil, false)),
		}, &Model{Path: "/a.model"}, false},
		{"withModelAttachment", &Decoder{
			p: newMockPackage(newMockFile("/a.model", []Relationship{{Type: RelType3DModel, Path: "/other.model"}}, otherModel, false)),
		}, &Model{Path: "/a.model", Childs: map[string]*ChildModel{"/other.model": new(ChildModel)}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := new(Model)
			_, err := tt.d.processOPC(model)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decoder.processOPC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(model, tt.want); diff != nil {
				t.Errorf("Decoder.processOPC() = %v", diff)
				return
			}
		})
	}
}

func TestDecoder_processRootModel_Fail(t *testing.T) {
	tests := []struct {
		name    string
		f       *mockFile
		wantErr bool
	}{
		{"errOpen", newMockFile("/a.model", nil, nil, true), true},
		{"invalidUnits", new(modelBuilder).withModel("other", "en-US", "").build(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := new(Decoder).processRootModel(context.Background(), tt.f, new(Model)); (err != nil) != tt.wantErr {
				t.Errorf("Decoder.processRootModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDecoder_processRootModel(t *testing.T) {
	baseMaterials := &BaseMaterials{ID: 5, Materials: []Base{
		{Name: "Blue PLA", Color: color.RGBA{0, 0, 255, 255}},
		{Name: "Red ABS", Color: color.RGBA{255, 0, 0, 255}},
	}}
	meshRes := &Object{
		Mesh: new(Mesh),
		ID:   8, Name: "Box 1", Thumbnail: "/a.png", PID: 5, PartNumber: "11111111-1111-1111-1111-111111111111",
	}
	meshRes.Mesh.Vertices = append(meshRes.Mesh.Vertices, []Point3D{
		{0, 0, 0},
		{100, 0, 0},
		{100, 100, 0},
		{0, 100, 0},
		{0, 0, 100},
		{100, 0, 100},
		{100, 100, 100},
		{0, 100, 100},
	}...)
	meshRes.Mesh.Triangles = append(meshRes.Mesh.Triangles, []Triangle{
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
	}...)

	components := &Object{
		ID: 20, Type: ObjectTypeSupport,
		Metadata:   []Metadata{{Name: xml.Name{Space: "qm", Local: "CustomMetadata3"}, Type: "xs:boolean", Value: "1"}, {Name: xml.Name{Space: "qm", Local: "CustomMetadata4"}, Type: "xs:boolean", Value: "2"}},
		Components: []*Component{{ObjectID: 8, Transform: Matrix{3, 0, 0, 0, 0, 1, 0, 0, 0, 0, 2, 0, -66.4, -87.1, 8.8, 1}}},
	}

	want := &Model{
		Units: UnitMillimeter, Language: "en-US", Path: "/3D/3dmodel.model", Thumbnail: "/thumbnail.png",
		Specs: map[string]spec.Spec{fakeExtension: &fakeSpec{}},
		Resources: Resources{
			Assets: []Asset{baseMaterials}, Objects: []*Object{meshRes, components},
		},
	}
	want.Build.Items = append(want.Build.Items, &Item{
		ObjectID: 20, PartNumber: "bob", Transform: Matrix{1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, -66.4, -87.1, 8.8, 1},
		Metadata: []Metadata{{Name: xml.Name{Space: "qm", Local: "CustomMetadata3"}, Type: "xs:boolean", Value: "1"}},
	})
	want.Metadata = append(want.Metadata, []Metadata{
		{Name: xml.Name{Local: "Application"}, Value: "go3mf app"},
		{Name: xml.Name{Space: "qm", Local: "CustomMetadata1"}, Preserve: true, Type: "xs:string", Value: "CE8A91FB-C44E-4F00-B634-BAA411465F6A"},
	}...)
	got := new(Model)
	got.Path = "/3D/3dmodel.model"
	rootFile := new(modelBuilder).withDefaultModel().withElement(`
		<resources>
			<basematerials id="5">
				<base name="Blue PLA" displaycolor="#0000FF" />
				<base name="Red ABS" displaycolor="#FF0000" />
			</basematerials>
			<object id="8" name="Box 1" pid="5" pindex="0" thumbnail="/a.png" partnumber="11111111-1111-1111-1111-111111111111" type="model">
				<mesh>
					<vertices>
						<vertex x="0" y="0" z="0" />
						<vertex x="100.00000" y="0" z="0" />
						<vertex x="100.00000" y="100.00000" z="0" />
						<vertex x="0" y="100.00000" z="0" />
						<vertex x="0" y="0" z="100.00000" />
						<vertex x="100.00000" y="0" z="100.00000" />
						<vertex x="100.00000" y="100.00000" z="100.00000" />
						<vertex x="0" y="100.00000" z="100.00000" />
					</vertices>
					<triangles>
						<triangle v1="3" v2="2" v3="1" />
						<triangle v1="1" v2="0" v3="3" />
						<triangle v1="4" v2="5" v3="6" p1="1" />
						<triangle v1="6" v2="7" v3="4" pid="5" p1="1" />
						<triangle v1="0" v2="1" v3="5" pid="5" p1="0" p2="1" p3="2"/>
						<triangle v1="5" v2="4" v3="0" pid="5" p1="3" p2="0" p3="2"/>
						<triangle v1="1" v2="2" v3="6" pid="5" p1="0" p2="1" p3="2"/>
						<triangle v1="6" v2="5" v3="1" pid="5" p1="2" p2="1" p3="3"/>
						<triangle v1="2" v2="3" v3="7" />
						<triangle v1="7" v2="6" v3="2" />
						<triangle v1="3" v2="0" v3="4" />
						<triangle v1="4" v2="7" v3="3" />
					</triangles>
				</mesh>
			</object>
			<object id="20" type="support">
				<metadatagroup>
					<metadata name="qm:CustomMetadata3" type="xs:boolean">1</metadata>
					<metadata name="qm:CustomMetadata4" type="xs:boolean">2</metadata>
				</metadatagroup>
				<components>
					<component objectid="8" transform="3 0 0 0 1 0 0 0 2 -66.4 -87.1 8.8"/>
				</components>
			</object>
		</resources>
		<build>
			<item partnumber="bob" objectid="20" transform="1 0 0 0 2 0 0 0 3 -66.4 -87.1 8.8">
				<metadatagroup>
					<metadata name="qm:CustomMetadata3" type="xs:boolean">1</metadata>
				</metadatagroup>
			</item>
		</build>
		<metadata name="Application">go3mf app</metadata>
		<metadata name="qm:CustomMetadata1" type="xs:string" preserve="1">CE8A91FB-C44E-4F00-B634-BAA411465F6A</metadata>
		<other />
		`).build("")

	d := new(Decoder)
	d.Strict = true
	got.WithSpec(&fakeSpec{})
	if err := d.processRootModel(context.Background(), rootFile, got); err != nil {
		t.Errorf("Decoder.processRootModel() unexpected error = %v", err)
		return
	}
	if diff := deep.Equal(got, want); diff != nil {
		t.Errorf("Decoder.processRootModel() = %v", diff)
		return
	}
}

func TestDecoder_processNonRootModels(t *testing.T) {
	tests := []struct {
		name    string
		model   *Model
		d       *Decoder
		wantErr bool
		want    *Model
	}{
		{"base", &Model{Childs: map[string]*ChildModel{"/3D/other.model": new(ChildModel), "/3D/new.model": new(ChildModel)}},
			&Decoder{nonRootModels: []packageFile{
				new(modelBuilder).withDefaultModel().withElement(`
				<resources>
					<basematerials id="5">
						<base name="Blue PLA" displaycolor="#0000FF" />
						<base name="Red ABS" displaycolor="#FF0000" />
					</basematerials>
				</resources>
			`).build("/3D/new.model"),
				new(modelBuilder).withDefaultModel().withElement(`
				<resources>
					<basematerials id="6" />
				</resources>
			`).build("/3D/other.model"),
			}}, false, &Model{
				Childs: map[string]*ChildModel{
					"/3D/other.model": {Resources: Resources{Assets: []Asset{&BaseMaterials{ID: 6}}}},
					"/3D/new.model": {Resources: Resources{Assets: []Asset{
						&BaseMaterials{ID: 5, Materials: []Base{
							{Name: "Blue PLA", Color: color.RGBA{0, 0, 255, 255}},
							{Name: "Red ABS", Color: color.RGBA{255, 0, 0, 255}},
						}}}}},
				},
			}},
		{"noAtt", new(Model), new(Decoder), false, new(Model)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.processNonRootModels(context.Background(), tt.model); (err != nil) != tt.wantErr {
				t.Errorf("Decoder.processNonRootModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(tt.model, tt.want); diff != nil {
				t.Errorf("Decoder.processNonRootModels() = %v", diff)
				return
			}
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name    string
		d       *Decoder
		wantErr bool
	}{
		{"base", &Decoder{
			p: newMockPackage(newMockFile("/a.model", []Relationship{{Type: "b", Path: "/a.xml"}}, nil, false)),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Decode(new(Model)); (err != nil) != tt.wantErr {
				t.Errorf("Decoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_modelFile_Decode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	checkEveryTokens = 1
	type args struct {
		ctx context.Context
		r   io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"nochild", args{context.Background(), bytes.NewBufferString(`
			<a></a>
			<b></b>
		`)}, false},
		{"eof", args{context.Background(), bytes.NewBufferString(`
			<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02">
				<build></build>
		`)}, true},
		{"canceled", args{ctx, bytes.NewBufferString(`
			<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02">
				<build></build>
			</model>
		`)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := decodeModelFile(tt.args.ctx, tt.args.r, new(Model), "", true, false); (err != nil) != tt.wantErr {
				t.Errorf("modelFile.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewDecoder(t *testing.T) {
	type args struct {
		r    io.ReaderAt
		size int64
	}
	tests := []struct {
		name string
		args args
		want *Decoder
	}{
		{"base", args{nil, 5}, &Decoder{
			Strict: true,
			p:      &opcReader{ra: nil, size: 5},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDecoder(tt.args.r, tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDecoder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecoder_processRootModel_warns(t *testing.T) {
	want := &specerr.List{Errors: []error{
		fmt.Errorf("Resources@BaseMaterials#0@RGBA#0: %v", specerr.NewParseAttrError("displaycolor", true)),
		fmt.Errorf("Resources@BaseMaterials#1: %v", specerr.NewParseAttrError("id", true)),
		fmt.Errorf("Resources@Object#0@Mesh@Point3D#8: %v", specerr.NewParseAttrError("x", true)),
		fmt.Errorf("Resources@Object#0@Mesh@Triangle#13: %v", specerr.NewParseAttrError("v1", true)),
		fmt.Errorf("Resources@Object#1: %v", specerr.NewParseAttrError("pid", false)),
		fmt.Errorf("Resources@Object#1: %v", specerr.NewParseAttrError("pindex", false)),
		fmt.Errorf("Resources@Object#1: %v", specerr.NewParseAttrError("type", false)),
		fmt.Errorf("Resources@Object#2@Component#0: %v", specerr.NewParseAttrError("transform", false)),
		fmt.Errorf("Resources@Object#2@Component#1: %v", specerr.NewParseAttrError("objectid", true)),
		fmt.Errorf("Build@Item#0: %v", specerr.NewParseAttrError("transform", false)),
		fmt.Errorf("Build@Item#3: %v", specerr.NewParseAttrError("objectid", true)),
	}}
	got := new(Model)
	got.Path = "/3D/3dmodel.model"
	rootFile := new(modelBuilder).withElement(`
		<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" 
		xmlns:qm="http://dummy.com/fake_ext" requiredextensions="qm other">
		<resources>
			<basematerials>
				<base name="Blue PLA" displaycolor="0000FF" />
				<base />
			</basematerials>
			<basematerials id="a"/>
			<basematerials id="5">
				<base name="Blue PLA" displaycolor="#0000FF" />
				<base name="Red ABS" displaycolor="#FF0000" />
			</basematerials>			
			<object id="8" name="Box 1" pid="5" pindex="0" partnumber="11111111-1111-1111-1111-111111111111" type="model">
				<mesh>
					<vertices>
						<vertex x="0" y="0" z="0" />
						<vertex x="100.00000" y="0" z="0" />
						<vertex x="100.00000" y="100.00000" z="0" />
						<vertex x="0" y="100.00000" z="0" />
						<vertex x="0" y="0" z="100.00000" />
						<vertex x="100.00000" y="0" z="100.00000" />
						<vertex x="100.00000" y="100.00000" z="100.00000" />
						<vertex x="0" y="100.00000" z="100.00000" />
						<vertex x="a" y="100.00000" z="100.00000" />
					</vertices>
					<triangles>
						<triangle v1="2" v2="2" v3="1" />
						<triangle v1="30" v2="2" v3="1" />
						<triangle v1="3" v2="2" v3="1" />
						<triangle v1="1" v2="0" v3="3" />
						<triangle v1="4" v2="5" v3="6" p1="1" />
						<triangle v1="6" v2="7" v3="4" pid="5" p1="1" />
						<triangle v1="0" v2="1" v3="5" pid="2" p1="0" p2="1" p3="2"/>
						<triangle v1="5" v2="4" v3="0" pid="2" p1="3" p2="0" p3="2"/>
						<triangle v1="1" v2="2" v3="6" pid="1" p1="0" p2="1" p3="2"/>
						<triangle v1="6" v2="5" v3="1" pid="1" p1="2" p2="1" p3="3"/>
						<triangle v1="2" v2="3" v3="7" />
						<triangle v1="7" v2="6" v3="2" />
						<triangle v1="3" v2="0" v3="4" />
						<triangle v1="a" v2="7" v3="3" />
					</triangles>
				</mesh>
			</object>
			<object id="22" pid="a" pindex="a" type="invalid" />
			<object id="20" pid="3" type="other">
				<components>
					<component objectid="8" transform="0 0 0 1 0 0 0 2 -66.4 -87.1 8.8"/>
					<component objectid="a"/>
				</components>
			</object>
		</resources>
		<build>
			<item partnumber="bob" objectid="20" transform="1 0 0 0 2 0 0 0 3 -66.4 -87.1" />
			<item objectid="8"/>
			<item objectid="5"/>
			<item objectid="a"/>
		</build>
		<metadata name="Application">go3mf app</metadata>
		<metadata name="qm:CustomMetadata1" type="xs:string" preserve="1">CE8A91FB-C44E-4F00-B634-BAA411465F6A</metadata>
		<metadata name="unknown:CustomMetadata1" type="xs:string" preserve="1">CE8A91FB-C44E-4F00-B634-BAA411465F6A</metadata>
		<other />
		</model>
		`).build("")

	t.Run("base", func(t *testing.T) {
		d := new(Decoder)
		d.Strict = false
		got.WithSpec(&fakeSpec{})
		err := d.processRootModel(context.Background(), rootFile, got)
		if diff := deep.Equal(err, want); diff != nil {
			t.Errorf("Decoder.processRootModel() = %v", diff)
			return
		}
	})
}

func TestOpenReader(t *testing.T) {
	r, err := OpenReader("testdata/cube.3mf")
	if err != nil {
		t.Errorf("OpenReader err = %v", err)
		return
	}
	defer r.Close()
	m := new(Model)
	err = r.Decode(m)
	if err != nil {
		t.Errorf("OpenReader.Decode err = %v", err)
		return
	}
	want := &Model{
		Language: "en-US", Path: "/3D/3dmodel.model",
		Resources: Resources{Objects: []*Object{
			{ID: 1, Name: "Cube", Mesh: &Mesh{
				Vertices: []Point3D{
					{100, 100, 100}, {100, 0, 100}, {100, 100, 0}, {0, 100, 0}, {100, 0, 0}, {}, {0, 0, 100}, {0, 100, 100},
				}, Triangles: []Triangle{
					NewTriangle(0, 1, 2), NewTriangle(3, 0, 2), NewTriangle(4, 3, 2),
					NewTriangle(5, 3, 4), NewTriangle(4, 6, 5), NewTriangle(6, 7, 5),
					NewTriangle(7, 6, 0), NewTriangle(1, 6, 4), NewTriangle(5, 7, 3),
					NewTriangle(7, 0, 3), NewTriangle(2, 1, 4), NewTriangle(0, 6, 1),
				}},
			},
		}},
		Build: Build{
			Items: []*Item{{
				ObjectID:  1,
				Transform: Identity().Translate(30, 30, 50),
			}},
		},
	}
	if diff := deep.Equal(m, want); diff != nil {
		t.Errorf("OpenReader.Decode() = %v", diff)
		return
	}
	if err = m.Validate(); err != nil {
		t.Errorf("OpenReader.Validate() err= %v", err)
		return
	}
}
