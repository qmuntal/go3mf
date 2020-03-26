package production

import (
	"github.com/qmuntal/go3mf"
	specerr "github.com/qmuntal/go3mf/errors"
)

func (e *Spec) ValidateAsset(_ *go3mf.Model, _ string, _ go3mf.Asset) []error {
	return nil
}

func (e *Spec) ValidateModel(m *go3mf.Model) []error {
	var (
		u    *UUID
		errs []error
	)
	if !m.Build.AnyAttr.Get(&u) {
		errs = append(errs, specerr.New(m.Build, &specerr.MissingFieldError{Name: attrProdUUID}))
	} else if validateUUID(string(*u)) != nil {
		errs = append(errs, specerr.New(m.Build, specerr.ErrUUID))
	}
	for i, item := range m.Build.Items {
		var iErrs []error
		var p *PathUUID
		if !item.AnyAttr.Get(&p) {
			iErrs = append(iErrs, &specerr.MissingFieldError{Name: attrProdUUID})
		} else {
			iErrs = e.validatePathUUID(m, "", p, iErrs)
		}
		for _, err := range iErrs {
			errs = append(errs, specerr.New(m.Build, specerr.NewIndexed(item, i, err)))
		}
	}
	return errs
}

func (e *Spec) ValidateObject(m *go3mf.Model, path string, obj *go3mf.Object) []error {
	var (
		u    *UUID
		errs []error
	)
	if !obj.AnyAttr.Get(&u) {
		errs = append(errs, &specerr.MissingFieldError{Name: attrProdUUID})
	} else if validateUUID(string(*u)) != nil {
		errs = append(errs, specerr.ErrUUID)
	}
	var p *PathUUID
	for i, c := range obj.Components {
		var cErrs []error
		if !c.AnyAttr.Get(&p) {
			cErrs = append(cErrs, &specerr.MissingFieldError{Name: attrProdUUID})
		} else {
			cErrs = e.validatePathUUID(m, path, p, cErrs)
		}
		for _, err := range cErrs {
			errs = append(errs, specerr.NewIndexed(c, i, err))
		}
	}
	return errs
}

func (e *Spec) validatePathUUID(m *go3mf.Model, path string, p *PathUUID, errs []error) []error {
	if p.UUID == "" {
		errs = append(errs, &specerr.MissingFieldError{Name: attrProdUUID})
	} else if validateUUID(string(p.UUID)) != nil {
		errs = append(errs, specerr.ErrUUID)
	}
	if p.Path != "" {
		if path == "" || path == m.PathOrDefault() { // root
			// Path is validated as part if the core validations
			if !e.Required() {
				errs = append(errs, specerr.ErrProdExtRequired)
			}
		} else {
			errs = append(errs, specerr.ErrProdRefInNonRoot)
		}
	}
	return errs
}