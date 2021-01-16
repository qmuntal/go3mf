package go3mf

import (
	"encoding/xml"
	"image/color"
	"sort"
	"strings"
	"sync"

	"github.com/qmuntal/go3mf/errors"
	"github.com/qmuntal/go3mf/spec"
)

func (m *Model) sortedSpecs() []string {
	if len(m.Specs) == 0 {
		return nil
	}
	s := make([]string, 0, len(m.Specs))
	for path := range m.Specs {
		s = append(s, path)
	}
	sort.Strings(s)
	return s
}

func (m *Model) sortedChilds() []string {
	if len(m.Childs) == 0 {
		return nil
	}
	s := make([]string, 0, len(m.Childs))
	for path := range m.Childs {
		s = append(s, path)
	}
	sort.Strings(s)
	return s
}

// Validate checks that the model is conformant with the 3MF specs.
func (m *Model) Validate() error {
	var errs error
	errs = errors.Append(errs, validateRelationship(m, m.RootRelationships, ""))
	errs = errors.Append(errs, m.validateNamespaces())
	rootPath := m.PathOrDefault()
	sortedChilds := m.sortedChilds()
	for _, path := range sortedChilds {
		c := m.Childs[path]
		if path == rootPath {
			errs = errors.Append(errs, errors.ErrOPCDuplicatedModelName)
		} else {
			errs = errors.Append(errs, validateRelationship(m, c.Relationships, path))
		}
	}

	errs = errors.Append(errs, validateRelationship(m, m.Relationships, rootPath))
	errs = errors.Append(errs, checkMetadadata(m, m.Metadata))

	sortedSpecs := m.sortedSpecs()
	for _, ns := range sortedSpecs {
		ext := m.Specs[ns]
		if ext, ok := ext.(spec.Validator); ok {
			errs = errors.Append(errs, ext.Validate(m, m.Path, m))
		}
	}

	for _, path := range sortedChilds {
		c := m.Childs[path]
		err := c.Resources.validate(m, path)
		if err != nil {
			errs = errors.Append(errs, errors.WrapPath(err, c.Resources, path))
		}
	}
	err := m.Resources.validate(m, rootPath)
	if err != nil {
		errs = errors.Append(errs, errors.Wrap(err, m.Resources))
	}
	err = m.Build.validate(m)
	if err != nil {
		errs = errors.Append(errs, errors.Wrap(err, m.Build))
	}
	return errs
}

func (item *Item) validate(m *Model) error {
	var errs error
	opath := item.ObjectPath()
	if item.ObjectID == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrObjectID))
	} else if obj, ok := m.FindObject(opath, item.ObjectID); ok {
		if obj.Type == ObjectTypeOther {
			errs = errors.Append(errs, errors.ErrOtherItem)
		}
	} else {
		errs = errors.Append(errs, errors.ErrMissingResource)
	}
	return errors.Append(errs, checkMetadadata(m, item.Metadata))
}

func (b *Build) validate(m *Model) error {
	var errs error
	for i, item := range b.Items {
		err := item.validate(m)
		if err != nil {
			errs = errors.Append(errs, errors.WrapIndex(err, item, i))
		}
	}
	return errs
}

var allowedMetadataNames = [...]string{ // sorted
	"application", "copyright", "creationdate", "description", "designer",
	"licenseterms", "modificationdate", "rating", "title",
}

func (m *Metadata) validate(model *Model) error {
	if m.Name.Local == "" {
		return errors.NewMissingFieldError(attrName)
	}
	var errs error
	if m.Name.Space == "" {
		nm := strings.ToLower(m.Name.Local)
		n := sort.SearchStrings(allowedMetadataNames[:], nm)
		if n >= len(allowedMetadataNames) || allowedMetadataNames[n] != nm {
			errs = errors.Append(errs, errors.ErrMetadataName)
		}
	} else {
		if _, ok := model.Specs[m.Name.Space]; !ok {
			errs = errors.Append(errs, errors.ErrMetadataNamespace)
		}
	}
	return errs
}

func checkMetadadata(model *Model, md []Metadata) error {
	var errs error
	names := make(map[xml.Name]struct{})
	for i, m := range md {
		err := m.validate(model)
		errs = errors.Append(errs, errors.WrapIndex(err, m, i))
		if _, ok := names[m.Name]; ok {
			errs = errors.Append(errs, errors.WrapIndex(errors.ErrMetadataDuplicated, m, i))
		}
		names[m.Name] = struct{}{}
	}
	return errs
}

// Validate validates the base materia is compliant with the 3MF specs.
func (r *BaseMaterials) Validate(m *Model, path string) error {
	var errs error
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if len(r.Materials) == 0 {
		errs = errors.Append(errs, errors.ErrEmptyResourceProps)
	}
	for j, b := range r.Materials {
		if b.Name == "" {
			errs = errors.Append(errs, errors.WrapIndex(errors.NewMissingFieldError(attrName), b, j))
		}
		if b.Color == (color.RGBA{}) {
			errs = errors.Append(errs, errors.WrapIndex(errors.NewMissingFieldError(attrDisplayColor), b, j))
		}
	}
	return errs
}

func (res *Resources) validate(m *Model, path string) error {
	var errs error
	assets := make(map[uint32]struct{})
	for i, r := range res.Assets {
		var aErrs error
		id := r.Identify()
		if id != 0 {
			if _, ok := assets[id]; ok {
				aErrs = errors.Append(aErrs, errors.ErrDuplicatedID)
			}
		}
		assets[id] = struct{}{}

		if r, ok := r.(*BaseMaterials); ok {
			aErrs = errors.Append(aErrs, r.Validate(m, path))
		}

		sortedSpecs := m.sortedSpecs()
		for _, ns := range sortedSpecs {
			ext := m.Specs[ns]
			if ext, ok := ext.(spec.Validator); ok {
				aErrs = errors.Append(aErrs, ext.Validate(m, path, r))
			}
		}
		errs = errors.Append(errs, errors.WrapIndex(aErrs, r, i))
	}
	for i, r := range res.Objects {
		if r.ID != 0 {
			if _, ok := assets[r.ID]; ok {
				errs = errors.Append(errs, errors.WrapIndex(errors.ErrDuplicatedID, r, i))
			}
		}
		assets[r.ID] = struct{}{}
		err := r.Validate(m, path)
		errs = errors.Append(errs, errors.WrapIndex(err, r, i))
	}
	return errs
}

// Validate validates that the object is compliant with 3MF specs,
// except for the mesh coherency.
func (r *Object) Validate(m *Model, path string) error {
	res, _ := m.FindResources(path)
	var errs error
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if r.PIndex != 0 && r.PID == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrPID))
	}
	if (r.Mesh != nil && len(r.Components) > 0) || (r.Mesh == nil && len(r.Components) == 0) {
		errs = errors.Append(errs, errors.ErrInvalidObject)
	}
	if r.Mesh != nil {
		if r.PID != 0 {
			if a, ok := res.FindAsset(r.PID); ok {
				if a, ok := a.(spec.PropertyGroup); ok {
					if int(r.PIndex) >= a.Len() {
						errs = errors.Append(errs, errors.ErrIndexOutOfBounds)
					}
				}
			} else {
				errs = errors.Append(errs, errors.ErrMissingResource)
			}
		}
		err := r.validateMesh(m, path)
		if err != nil {
			errs = errors.Append(errs, errors.Wrap(err, r.Mesh))
		}
	}
	if len(r.Components) > 0 {
		if r.PID != 0 {
			errs = errors.Append(errs, errors.ErrComponentsPID)
		}
		errs = errors.Append(errs, r.validateComponents(m, path))
	}

	sortedSpecs := m.sortedSpecs()
	for _, ns := range sortedSpecs {
		ext := m.Specs[ns]
		if ext, ok := ext.(spec.Validator); ok {
			errs = errors.Append(errs, ext.Validate(m, path, r))
		}
	}
	return errs
}

func (r *Object) validateMesh(m *Model, path string) error {
	res, _ := m.FindResources(path)
	var errs error
	switch r.Type {
	case ObjectTypeModel, ObjectTypeSolidSupport:
		if len(r.Mesh.Vertices) < 3 {
			errs = errors.Append(errs, errors.ErrInsufficientVertices)
		}
		if len(r.Mesh.Triangles) <= 3 && len(r.Mesh.Any) == 0 {
			errs = errors.Append(errs, errors.ErrInsufficientTriangles)
		}
	}

	nodeCount := uint32(len(r.Mesh.Vertices))
	for i, face := range r.Mesh.Triangles {
		i0, i1, i2 := face.Indices()
		if i0 == i1 || i0 == i2 || i1 == i2 {
			errs = errors.Append(errs, errors.WrapIndex(errors.ErrDuplicatedIndices, face, i))
		}
		if i0 >= nodeCount || i1 >= nodeCount || i2 >= nodeCount {
			errs = errors.Append(errs, errors.WrapIndex(errors.ErrIndexOutOfBounds, face, i))
		}
		pid := face.PID()
		if pid != 0 {
			p1, p2, p3 := face.PIndices()
			if pid == r.PID && p1 == r.PIndex &&
				p2 == r.PIndex && p3 == r.PIndex {
				continue
			}
			if a, ok := res.FindAsset(pid); ok {
				if a, ok := a.(spec.PropertyGroup); ok {
					l := a.Len()
					if int(p1) >= l || int(p2) >= l || int(p3) >= l {
						errs = errors.Append(errs, errors.WrapIndex(errors.ErrIndexOutOfBounds, face, i))
					}
				}
			} else {
				errs = errors.Append(errs, errors.WrapIndex(errors.ErrMissingResource, face, i))
			}
		}
	}
	return errs
}

func (r *Object) validateComponents(m *Model, path string) error {
	var errs error
	for j, c := range r.Components {
		if c.ObjectID == 0 {
			errs = errors.Append(errs, errors.WrapIndex(errors.NewMissingFieldError(attrObjectID), c, j))
		} else if ref, ok := m.FindObject(c.ObjectPath(path), c.ObjectID); ok {
			if ref.ID == r.ID && c.ObjectPath(path) == path {
				errs = errors.Append(errs, errors.WrapIndex(errors.ErrRecursion, c, j))
			}
		} else {
			errs = errors.Append(errs, errors.WrapIndex(errors.ErrMissingResource, c, j))
		}
	}
	return errs
}

func (m *Model) validateNamespaces() error {
	sortedSpecs := m.sortedSpecs()
	for _, ns := range sortedSpecs {
		ext := m.Specs[ns]
		if ext.Required() {
			if _, ok := ext.(*spec.UnknownSpec); ok {
				return errors.ErrRequiredExt
			}
		}
	}
	return nil
}

func validateRelationship(m *Model, rels []Relationship, path string) error {
	var errs error
	type partrel struct{ path, rel string }
	visitedParts := make(map[partrel]struct{})
	var hasPrintTicket bool
	for i, r := range rels {
		if r.Path == "" || r.Path[0] != '/' || strings.Contains(r.Path, "/.") {
			errs = errors.Append(errs, errors.WrapIndex(errors.ErrOPCPartName, r, i))
		} else {
			if _, ok := findAttachment(m.Attachments, r.Path); !ok {
				errs = errors.Append(errs, errors.WrapIndex(errors.ErrOPCRelTarget, r, i))
			}
			if _, ok := visitedParts[partrel{r.Path, r.Type}]; ok {
				errs = errors.Append(errs, errors.WrapIndex(errors.ErrOPCDuplicatedRel, r, i))
			}
			visitedParts[partrel{r.Path, r.Type}] = struct{}{}
		}
		switch r.Type {
		case RelTypePrintTicket:
			if a, ok := findAttachment(m.Attachments, r.Path); ok {
				if a.ContentType != ContentTypePrintTicket {
					errs = errors.Append(errs, errors.WrapIndex(errors.ErrOPCContentType, r, i))
				}
				if hasPrintTicket {
					errs = errors.Append(errs, errors.WrapIndex(errors.ErrOPCDuplicatedTicket, r, i))
				}
				hasPrintTicket = true
			}
		}
	}
	if path != "" && errs != nil {
		for _, err := range errs.(*errors.List).Errors {
			if err, ok := err.(*errors.Error); ok {
				err.Path = path
			}
		}
	}
	return errs
}

func findAttachment(att []Attachment, path string) (*Attachment, bool) {
	for _, a := range att {
		if strings.EqualFold(a.Path, path) {
			return &a, true
		}
	}
	return nil, false
}

// ValidateCoherency checks that all the mesh are non-empty, manifold and oriented.
func (m *Model) ValidateCoherency() error {
	var (
		errs error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)
	wg.Add(len(m.Resources.Objects))
	for i := range m.Resources.Objects {
		go func(i int) {
			defer wg.Done()
			r := m.Resources.Objects[i]
			if isSolidObject(r) {
				err := r.Mesh.ValidateCoherency()
				if err != nil {
					mu.Lock()
					errs = errors.Append(errs, errors.Wrap(errors.WrapIndex(errors.Wrap(err, r.Mesh), r, i), m.Resources))
					mu.Unlock()
				}
			}
		}(i)
	}
	for path, c := range m.Childs {
		wg.Add(len(c.Resources.Objects))
		for i := range c.Resources.Objects {
			go func(i int) {
				defer wg.Done()
				r := c.Resources.Objects[i]
				if isSolidObject(r) {
					err := r.Mesh.ValidateCoherency()
					if err != nil {
						mu.Lock()
						errs = errors.Append(errs, errors.WrapPath(errors.WrapIndex(errors.Wrap(err, r.Mesh), r, i), c.Resources, path))
						mu.Unlock()
					}
				}
			}(i)
		}
	}
	wg.Wait()
	return errs
}

func isSolidObject(r *Object) bool {
	return r.Mesh != nil && (r.Type == ObjectTypeModel || r.Type == ObjectTypeSolidSupport)
}

// ValidateCoherency checks that the mesh is non-empty, manifold and oriented.
func (m *Mesh) ValidateCoherency() error {
	if len(m.Vertices) < 3 {
		return errors.ErrInsufficientVertices
	}
	if len(m.Triangles) <= 3 {
		return errors.ErrInsufficientTriangles
	}

	var edgeCounter uint32
	pairMatching := make(pairMatch)
	for _, face := range m.Triangles {
		for j := 0; j < 3; j++ {
			n1, n2 := face[j].ToUint32(), face[(j+1)%3].ToUint32()
			if _, ok := pairMatching.CheckMatch(n1, n2); !ok {
				pairMatching.AddMatch(n1, n2, edgeCounter)
				edgeCounter++
			}
		}
	}

	positive, negative := make([]uint32, edgeCounter), make([]uint32, edgeCounter)
	for _, face := range m.Triangles {
		for j := 0; j < 3; j++ {
			n1, n2 := face[j].ToUint32(), face[(j+1)%3].ToUint32()
			edgeIndex, _ := pairMatching.CheckMatch(n1, n2)
			if n1 <= n2 {
				positive[edgeIndex]++
			} else {
				negative[edgeIndex]++
			}
		}
	}

	for i := uint32(0); i < edgeCounter; i++ {
		if positive[i] != 1 || negative[i] != 1 {
			return errors.ErrMeshConsistency
		}
	}
	return nil
}
