package materials

import (
	"image/color"
	"strings"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/errors"
)

func (sp *Spec) Validate(path string, asset interface{}) error {
	if asset, ok := asset.(go3mf.Asset); ok {
		return sp.validateAsset(path, asset)
	}
	return nil
}

func (sp *Spec) validateAsset(path string, r go3mf.Asset) (errs error) {
	switch r := r.(type) {
	case *ColorGroup:
		errs = sp.validateColorGroup(path, r)
	case *Texture2DGroup:
		errs = sp.validateTexture2DGroup(path, r)
	case *Texture2D:
		errs = sp.validateTexture2D(path, r)
	case *MultiProperties:
		errs = sp.validateMultiProps(path, r)
	case *CompositeMaterials:
		errs = sp.validateCompositeMat(path, r)
	}
	return
}

func (sp *Spec) validateColorGroup(path string, r *ColorGroup) (errs error) {
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if len(r.Colors) == 0 {
		errs = errors.Append(errs, errors.ErrEmptyResourceProps)
	}
	for j, c := range r.Colors {
		if c == (color.RGBA{}) {
			errs = errors.Append(errs, errors.WrapIndex(errors.NewMissingFieldError(attrColor), c, j))
		}
	}
	return
}

func (sp *Spec) validateTexture2DGroup(path string, r *Texture2DGroup) (errs error) {
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if r.TextureID == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrTexID))
	} else if text, ok := sp.m.FindAsset(path, r.TextureID); ok {
		if _, ok := text.(*Texture2D); !ok {
			errs = errors.Append(errs, ErrTextureReference)
		}
	} else {
		errs = errors.Append(errs, ErrTextureReference)
	}
	if len(r.Coords) == 0 {
		errs = errors.Append(errs, errors.ErrEmptyResourceProps)
	}
	return
}

func (sp *Spec) validateTexture2D(path string, r *Texture2D) (errs error) {
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if r.Path == "" {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrPath))
	} else {
		var hasTexture bool
		for _, a := range sp.m.Attachments {
			if strings.EqualFold(a.Path, r.Path) {
				hasTexture = true
				break
			}
		}
		if !hasTexture {
			errs = errors.Append(errs, ErrMissingTexturePart)
		}
	}
	if r.ContentType == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrContentType))
	}
	return
}

func (sp *Spec) validateMultiProps(path string, r *MultiProperties) (errs error) {
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if len(r.PIDs) == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrPIDs))
	}
	if len(r.BlendMethods) > len(r.PIDs)-1 {
		errs = errors.Append(errs, ErrMultiBlend)
	}
	if len(r.Multis) == 0 {
		errs = errors.Append(errs, errors.ErrEmptyResourceProps)
	}
	var (
		colorCount        int
		resourceUndefined bool
		lengths           = make([]int, len(r.PIDs))
	)
	for j, pid := range r.PIDs {
		if pr, ok := sp.m.FindAsset(path, pid); ok {
			switch pr := pr.(type) {
			case *go3mf.BaseMaterials:
				if j != 0 {
					errs = errors.Append(errs, ErrMaterialMulti)
				}
				lengths[j] = len(pr.Materials)
			case *CompositeMaterials:
				if j != 0 {
					errs = errors.Append(errs, ErrMaterialMulti)
				}
				lengths[j] = len(pr.Composites)
			case *MultiProperties:
				errs = errors.Append(errs, ErrMultiRefMulti)
			case *ColorGroup:
				if colorCount == 1 {
					errs = errors.Append(errs, ErrMultiColors)
				}
				colorCount++
				lengths[j] = len(pr.Colors)
			}
		} else if !resourceUndefined {
			resourceUndefined = true
			errs = errors.Append(errs, errors.ErrMissingResource)
		}
	}
	for j, m := range r.Multis {
		for k, index := range m.PIndices {
			if k < len(r.PIDs) && lengths[k] < int(index) {
				errs = errors.Append(errs, errors.WrapIndex(errors.ErrIndexOutOfBounds, m, j))
				break
			}
		}
	}
	return
}

func (sp *Spec) validateCompositeMat(path string, r *CompositeMaterials) (errs error) {
	if r.ID == 0 {
		errs = errors.Append(errs, errors.ErrMissingID)
	}
	if r.MaterialID == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrMatID))
	} else if mat, ok := sp.m.FindAsset(path, r.MaterialID); ok {
		if bm, ok := mat.(*go3mf.BaseMaterials); ok {
			for _, index := range r.Indices {
				if int(index) > len(bm.Materials) {
					errs = errors.Append(errs, errors.ErrIndexOutOfBounds)
					break
				}
			}
		} else {
			errs = errors.Append(errs, ErrCompositeBase)
		}
	} else {
		errs = errors.Append(errs, errors.ErrMissingResource)
	}
	if len(r.Indices) == 0 {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrMatIndices))
	}
	if len(r.Composites) == 0 {
		errs = errors.Append(errs, errors.ErrEmptyResourceProps)
	}
	return
}
