package production

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/errors"
)

func TestDecode(t *testing.T) {
	components := &go3mf.Object{
		AnyAttr: go3mf.AnyAttr{&ObjectAttr{UUID: "cb828680-8895-4e08-a1fc-be63e033df15"}},
		ID:      20,
		Components: []*go3mf.Component{{
			AnyAttr: go3mf.AnyAttr{&ComponentAttr{
				Path: "/3D/other.model",
				UUID: "cb828680-8895-4e08-a1fc-be63e033df16",
			}},
			ObjectID: 8, Transform: go3mf.Matrix{3, 0, 0, 0, 0, 1, 0, 0, 0, 0, 2, 0, -66.4, -87.1, 8.8, 1}},
		},
	}

	want := &go3mf.Model{Path: "/3D/3dmodel.model", Resources: go3mf.Resources{
		Objects: []*go3mf.Object{components},
	}, Build: go3mf.Build{
		AnyAttr: go3mf.AnyAttr{&BuildAttr{UUID: "e9e25302-6428-402e-8633-cc95528d0ed3"}}},
	}
	want.Build.Items = append(want.Build.Items, &go3mf.Item{ObjectID: 20,
		AnyAttr:   go3mf.AnyAttr{&ItemAttr{UUID: "e9e25302-6428-402e-8633-cc95528d0ed2"}},
		Transform: go3mf.Matrix{1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, -66.4, -87.1, 8.8, 1},
	}, &go3mf.Item{ObjectID: 8,
		AnyAttr: go3mf.AnyAttr{&ItemAttr{
			Path: "/3D/other.model",
			UUID: "e9e25302-6428-402e-8633-cc95528d0ed4",
		}},
	})
	got := new(go3mf.Model)
	got.Path = "/3D/3dmodel.model"
	rootFile := `
		<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" xmlns:p="http://schemas.microsoft.com/3dmanufacturing/production/2015/06">
		<resources>
			<object id="20" p:UUID="cb828680-8895-4e08-a1fc-be63e033df15">
				<components>
					<component objectid="8" p:UUID="cb828680-8895-4e08-a1fc-be63e033df16" p:path="/3D/other.model" transform="3 0 0 0 1 0 0 0 2 -66.4 -87.1 8.8"/>
				</components>
			</object>
		</resources>
		<build p:UUID="e9e25302-6428-402e-8633-cc95528d0ed3">
			<item objectid="20" p:UUID="e9e25302-6428-402e-8633-cc95528d0ed2" transform="1 0 0 0 2 0 0 0 3 -66.4 -87.1 8.8" />
			<item objectid="8" p:UUID="e9e25302-6428-402e-8633-cc95528d0ed4" p:path="/3D/other.model" />
		</build>
		</model>
		`
	t.Run("base", func(t *testing.T) {
		want.WithSpec(&Spec{LocalName: "p"})
		got.WithSpec(&Spec{LocalName: "p"})
		if err := go3mf.UnmarshalModel([]byte(rootFile), got); err != nil {
			t.Errorf("UnmarshalModel() unexpected error = %v", err)
			return
		}
		if diff := deep.Equal(got, want); diff != nil {
			t.Errorf("UnmarshalModel() = %v", diff)
			return
		}
	})
}

func TestDecode_warns(t *testing.T) {
	want := &errors.List{Errors: []error{
		fmt.Errorf("Resources@Object#1: %v", &errors.ParseAttrError{Required: true, Name: "UUID"}),
		fmt.Errorf("Resources@Object#1@Component#0: %v", &errors.ParseAttrError{Required: true, Name: "UUID"}),
		fmt.Errorf("Build: %v", &errors.ParseAttrError{Required: true, Name: "UUID"}),
		fmt.Errorf("Build@Item#0: %v", &errors.ParseAttrError{Required: true, Name: "UUID"}),
	}}
	got := new(go3mf.Model)
	got.Path = "/3D/3dmodel.model"
	rootFile := `
		<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" xmlns:p="http://schemas.microsoft.com/3dmanufacturing/production/2015/06">
		<resources>
			<object id="22" p:UUID="cb828680-8895-4e08-a1fc-be63e033df15" />
			<object id="20" p:UUID="cb8286808895-4e08-a1fc-be63e033df15">
				<components>
					<component objectid="8" p:path="/2d/2d.model" p:UUID="cb8286808895-4e08-a1fc-be63e033df16"/>
					<component objectid="5" p:UUID="cb828680-8895-4e08-a1fc-be63e033df16"/>
				</components>
			</object>
		</resources>
		<build p:UUID="e9e25302-6428-402e-8633ed2">
			<item partnumber="bob" objectid="20" p:UUID="invalid-uuid" />
			<item objectid="8" p:path="/3D/other.model"/>
			<item objectid="5" p:UUID="e9e25302-6428-402e-8633-cc95528d0ed4"/>
		</build>
		</model>`

	t.Run("base", func(t *testing.T) {
		got.WithSpec(&Spec{LocalName: "p"})
		err := go3mf.UnmarshalModel([]byte(rootFile), got)
		if diff := deep.Equal(err, want); diff != nil {
			t.Errorf("UnmarshalModel_warn() = %v", diff)
			return
		}
	})
}
