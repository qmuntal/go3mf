package production

import (
	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/errors"
	"github.com/qmuntal/go3mf/uuid"
)

type uuidPath interface {
	getUUID() string
	ObjectPath() string
}

func (sp *Spec) Validate(path string, e interface{}) error {
	switch e := e.(type) {
	case *go3mf.Model:
		return sp.validateModel(e)
	case *go3mf.Object:
		return sp.validateObject(path, e)
	}
	return nil
}

func (sp *Spec) validateModel(m *go3mf.Model) error {
	var errs error
	u := GetBuildAttr(&m.Build)
	if u == nil {
		errs = errors.Append(errs, errors.Wrap(errors.NewMissingFieldError(attrProdUUID), m.Build))
	} else if uuid.Validate(u.UUID) != nil {
		errs = errors.Append(errs, errors.Wrap(ErrUUID, m.Build))
	}
	for i, item := range m.Build.Items {
		var iErrs error

		if p := GetItemAttr(item); p != nil {
			iErrs = errors.Append(iErrs, sp.validatePathUUID("", p))
		} else {
			iErrs = errors.Append(iErrs, errors.NewMissingFieldError(attrProdUUID))
		}
		if iErrs != nil {
			errs = errors.Append(errs, errors.Wrap(errors.WrapIndex(iErrs, item, i), m.Build))
		}
	}
	return errs
}

func (sp *Spec) validateObject(path string, obj *go3mf.Object) error {
	var errs error
	u := GetObjectAttr(obj)
	if u == nil {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrProdUUID))
	} else if uuid.Validate(u.UUID) != nil {
		errs = errors.Append(errs, ErrUUID)
	}
	for i, c := range obj.Components {
		var cErrs error
		if p := GetComponentAttr(c); p != nil {
			cErrs = errors.Append(cErrs, sp.validatePathUUID(path, p))
		} else {
			cErrs = errors.Append(cErrs, errors.NewMissingFieldError(attrProdUUID))
		}
		if cErrs != nil {
			errs = errors.Append(errs, errors.WrapIndex(cErrs, c, i))
		}
	}
	return errs
}

func (sp *Spec) validatePathUUID(path string, p uuidPath) error {
	var errs error
	if p.getUUID() == "" {
		errs = errors.Append(errs, errors.NewMissingFieldError(attrProdUUID))
	} else if uuid.Validate(string(p.getUUID())) != nil {
		errs = errors.Append(errs, ErrUUID)
	}
	if p.ObjectPath() != "" {
		if path == "" || path == sp.m.PathOrDefault() { // root
			// Path is validated as part if the core validations
		} else {
			errs = errors.Append(errs, ErrProdRefInNonRoot)
		}
	}
	return errs
}
