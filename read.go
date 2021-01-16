package go3mf

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"unsafe"

	specerr "github.com/qmuntal/go3mf/errors"
	xml3mf "github.com/qmuntal/go3mf/internal/xml"
	"github.com/qmuntal/go3mf/spec/encoding"
)

var checkEveryTokens = 1000

type packageFile interface {
	Name() string
	ContentType() string
	FindFileFromName(string) (packageFile, bool)
	Relationships() []Relationship
	Open() (io.ReadCloser, error)
}

type packageReader interface {
	Open(func(r io.Reader) io.ReadCloser) error
	FindFileFromName(string) (packageFile, bool)
	Relationships() []Relationship
}

// ReadCloser wrapps a Decoder than can be closed.
type ReadCloser struct {
	f *os.File
	*Decoder
}

// OpenReader will open the 3MF file specified by name and return a ReadCloser.
func OpenReader(name string) (*ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &ReadCloser{f: f, Decoder: NewDecoder(f, fi.Size())}, nil
}

// Close closes the 3MF file, rendering it unusable for I/O.
func (r *ReadCloser) Close() error {
	return r.f.Close()
}

func decodeModelFile(ctx context.Context, r io.Reader, model *Model, path string, isRoot, strict bool) error {
	x := xml3mf.NewDecoder(r)
	scanner := decoderContext{
		extensionDecoder: make(map[string]encoding.Decoder),
		isRoot:           isRoot,
		modelPath:        path,
	}
	for _, ext := range model.Specs {
		if ext, ok := ext.(encoding.Decoder); ok {
			scanner.extensionDecoder[ext.Namespace()] = ext
		}
	}
	state, names := make([]encoding.ElementDecoder, 0, 10), make([]xml.Name, 0, 10)

	var (
		currentDecoder, tmpDecoder encoding.ElementDecoder
		currentName                xml.Name
	)
	currentDecoder = &topLevelDecoder{isRoot: isRoot, model: model, ctx: &scanner}
	var err error
	x.OnStart = func(tp xml3mf.StartElement) {
		if childDecoder, ok := currentDecoder.(encoding.ChildElementDecoder); ok {
			tmpDecoder = childDecoder.Child(tp.Name)
			if tmpDecoder != nil {
				state = append(state, currentDecoder)
				names = append(names, currentName)
				currentName = tp.Name
				currentDecoder = tmpDecoder
				if err := currentDecoder.Start(*(*[]encoding.XMLAttr)(unsafe.Pointer(&tp.Attr))); err != nil {
					if childDecoder, ok := childDecoder.(encoding.ErrorWrapper); ok {
						err = childDecoder.Wrap(err)
					}
					specerr.Append(&scanner.Err, err)
				}
			}
		}
	}
	x.OnEnd = func(tp xml.EndElement) {
		if currentName == tp.Name {
			currentDecoder.End()
			currentDecoder, state = state[len(state)-1], state[:len(state)-1]
			currentName, names = names[len(names)-1], names[:len(names)-1]
		}
	}
	x.OnChar = func(tp xml.CharData) {
		if currentDecoder, ok := currentDecoder.(encoding.CharDataElementDecoder); ok {
			currentDecoder.CharData(tp)
		}
	}
	var i int
	for {
		err = x.RawToken()
		if err != nil || (strict && scanner.Err.Len() != 0) {
			break
		}
		if i%checkEveryTokens == 0 {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			default: // Default is must to avoid blocking
			}
			if err != nil {
				break
			}
		}
		i++
	}
	if err == io.EOF {
		err = nil
	}
	if err == nil && scanner.Err.Len() != 0 {
		if strict || scanner.Err.Len() == 1 {
			err = scanner.Err.Unwrap()
		} else {
			err = &scanner.Err
		}
	}
	return err
}

// Decoder implements a 3mf file decoder.
type Decoder struct {
	Strict        bool
	p             packageReader
	flate         func(r io.Reader) io.ReadCloser
	nonRootModels []packageFile
}

// NewDecoder returns a new Decoder reading a 3mf file from r.
func NewDecoder(r io.ReaderAt, size int64) *Decoder {
	return &Decoder{
		p:      &opcReader{ra: r, size: size},
		Strict: true,
	}
}

// Decode reads the 3mf file and unmarshall its content into the model.
func (d *Decoder) Decode(model *Model) error {
	return d.DecodeContext(context.Background(), model)
}

// DecodeContext reads the 3mf file and unmarshall its content into the model.
func (d *Decoder) DecodeContext(ctx context.Context, model *Model) error {
	rootFile, err := d.processOPC(model)
	if err != nil {
		return err
	}
	if err := d.processNonRootModels(ctx, model); err != nil {
		return err
	}
	return d.processRootModel(ctx, rootFile, model)
}

// UnmarshalModel fills a model with the data of a root model file
// using not strict mode.
func UnmarshalModel(data []byte, model *Model) error {
	d := NewDecoder(nil, 0)
	d.Strict = false
	return d.processRootModel(context.Background(), &fakePackageFile{data: data}, model)
}

func (d *Decoder) processRootModel(ctx context.Context, rootFile packageFile, model *Model) error {
	f, err := rootFile.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	err = decodeModelFile(ctx, f, model, rootFile.Name(), true, d.Strict)
	if err != nil {
		return err
	}
	return nil
}

func (d *Decoder) processNonRootModels(ctx context.Context, model *Model) (errs error) {
	var (
		wg                 sync.WaitGroup
		nonRootModelsCount = len(d.nonRootModels)
	)
	wg.Add(nonRootModelsCount)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for i := 0; i < nonRootModelsCount; i++ {
		go func(i int) {
			defer wg.Done()
			err := d.readChildModel(ctx, i, model)
			if err != nil {
				errs = err
				cancel()
			}
		}(i)
	}
	wg.Wait()
	if errs != nil {
		return errs
	}
	return nil
}

func (d *Decoder) processOPC(model *Model) (packageFile, error) {
	if err := d.p.Open(d.flate); err != nil {
		return nil, err
	}
	var rootFile packageFile
	for _, r := range d.p.Relationships() {
		if r.Type == RelType3DModel {
			var ok bool
			rootFile, ok = d.p.FindFileFromName(r.Path)
			if !ok {
				return nil, errors.New("package root model points to an unexisting file")
			}
			model.Path = rootFile.Name()
			d.extractCoreAttachments(rootFile, model, true)
			for _, file := range d.nonRootModels {
				d.extractCoreAttachments(file, model, false)
			}
		} else if att, ok := d.p.FindFileFromName(r.Path); ok {
			model.RootRelationships = append(model.RootRelationships, r)
			model.Attachments = d.addAttachment(model.Attachments, att)
		}
	}
	if rootFile == nil {
		return nil, errors.New("package does not have root model")
	}
	return rootFile, nil
}

func (d *Decoder) extractCoreAttachments(modelFile packageFile, model *Model, isRoot bool) {
	for _, rel := range modelFile.Relationships() {
		if file, ok := modelFile.FindFileFromName(rel.Path); ok {
			if isRoot {
				if rel.Type == RelType3DModel {
					d.nonRootModels = append(d.nonRootModels, file)
					if model.Childs == nil {
						model.Childs = make(map[string]*ChildModel)
					}
					model.Childs[file.Name()] = new(ChildModel)
				} else {
					model.Attachments = d.addAttachment(model.Attachments, file)
					model.Relationships = append(model.Relationships, rel)
				}
			} else if rel.Type != RelType3DModel {
				if child, ok := model.Childs[modelFile.Name()]; ok {
					model.Attachments = d.addAttachment(model.Attachments, file)
					child.Relationships = append(child.Relationships, rel)
				}
			}
		}
	}
}

func (d *Decoder) addAttachment(attachments []Attachment, file packageFile) []Attachment {
	for _, att := range attachments {
		if strings.EqualFold(att.Path, file.Name()) {
			return attachments
		}
	}
	if buff, err := copyFile(file); err == nil {
		return append(attachments, Attachment{
			Path:        file.Name(),
			Stream:      buff,
			ContentType: file.ContentType(),
		})
	}
	return attachments
}

func (d *Decoder) readChildModel(ctx context.Context, i int, model *Model) error {
	attachment := d.nonRootModels[i]
	file, err := attachment.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	err = decodeModelFile(ctx, file, model, attachment.Name(), false, d.Strict)
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default: // Default is must to avoid blocking
	}
	return err
}

func copyFile(file packageFile) (io.Reader, error) {
	stream, err := file.Open()
	if err != nil {
		return nil, err
	}
	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, stream)
	stream.Close()
	return buff, err
}

type fakePackageFile struct {
	data []byte
}

func (f *fakePackageFile) Name() string                                { return DefaultModelPath }
func (f *fakePackageFile) ContentType() string                         { return ContentType3DModel }
func (f *fakePackageFile) FindFileFromName(string) (packageFile, bool) { return nil, false }
func (f *fakePackageFile) Relationships() []Relationship               { return nil }
func (f *fakePackageFile) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBuffer(f.data)), nil
}

// A decoderContext is a 3mf model file scanning state machine.
type decoderContext struct {
	Err              specerr.List
	extensionDecoder map[string]encoding.Decoder
	modelPath        string
	isRoot           bool
}

func (s *decoderContext) namespace(local string) (string, bool) {
	for _, ext := range s.extensionDecoder {
		if ext.Local() == local {
			return ext.Namespace(), true
		}
	}
	return "", false
}
