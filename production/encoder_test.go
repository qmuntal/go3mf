package production

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/qmuntal/go3mf"
)

func TestMarshalModel(t *testing.T) {
	components := &go3mf.Object{
		AnyAttr: go3mf.AnyAttr{&ObjectAttr{UUID: "cb828680-8895-4e08-a1fc-be63e033df15"}},
		ID:      20,
		Components: []*go3mf.Component{{
			ObjectID: 8,
			AnyAttr: go3mf.AnyAttr{&ComponentAttr{
				Path: "/3D/other.model",
				UUID: "cb828680-8895-4e08-a1fc-be63e033df16",
			}}},
		},
	}
	m := &go3mf.Model{Path: "/3D/3dmodel.model", Build: go3mf.Build{
		AnyAttr: go3mf.AnyAttr{&BuildAttr{UUID: "e9e25302-6428-402e-8633-cc95528d0ed3"}},
	}}
	m.Resources = go3mf.Resources{Objects: []*go3mf.Object{components}}
	m.Build.Items = append(m.Build.Items, &go3mf.Item{ObjectID: 20,
		AnyAttr: go3mf.AnyAttr{&ItemAttr{UUID: "e9e25302-6428-402e-8633-cc95528d0ed2"}},
	}, &go3mf.Item{ObjectID: 8,
		AnyAttr: go3mf.AnyAttr{&ItemAttr{
			Path: "/3D/other.model",
			UUID: "e9e25302-6428-402e-8633-cc95528d0ed4",
		}},
	})
	m.WithSpec(&Spec{LocalName: "p"})
	b, err := go3mf.MarshalModel(m)
	if err != nil {
		t.Errorf("production.MarshalModel() error = %v", err)
		return
	}
	newModel := new(go3mf.Model)
	newModel.WithSpec(&Spec{LocalName: "p"})
	newModel.Path = m.Path
	if err := go3mf.UnmarshalModel(b, newModel); err != nil {
		t.Errorf("production.MarshalModel() error decoding = %v, s = %s", err, string(b))
		return
	}
	if diff := deep.Equal(m, newModel); diff != nil {
		t.Errorf("production.MarshalModel() = %v, s = %s", diff, string(b))
	}
}
