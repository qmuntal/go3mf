package production

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/errors"
)

func TestValidate(t *testing.T) {
	validMesh := &go3mf.Object{ID: 1, Mesh: &go3mf.Mesh{Vertices: []go3mf.Point3D{{}, {}, {}, {}}, Triangles: []go3mf.Triangle{
		go3mf.NewTriangle(0, 1, 2), go3mf.NewTriangle(0, 3, 1), go3mf.NewTriangle(0, 2, 3), go3mf.NewTriangle(1, 3, 2),
	}}}
	tests := []struct {
		name  string
		model *go3mf.Model
		want  []error
	}{
		{"buildNoUUID", &go3mf.Model{Build: go3mf.Build{}}, []error{
			fmt.Errorf("Build: %v", &errors.MissingFieldError{Name: attrProdUUID}),
		}},
		{"buildEmptyUUID", &go3mf.Model{Build: go3mf.Build{
			AnyAttr: go3mf.AnyAttr{&BuildAttr{}}}}, []error{
			fmt.Errorf("Build: %v", ErrUUID),
		}},
		{"buildNonValidUUID", &go3mf.Model{Build: go3mf.Build{
			AnyAttr: go3mf.AnyAttr{&BuildAttr{"a-b-c-d"}}}}, []error{
			fmt.Errorf("Build: %v", ErrUUID),
		}},
		{"extReq", &go3mf.Model{Specs: map[string]go3mf.Spec{Namespace: &Spec{}},
			Childs: map[string]*go3mf.ChildModel{"/other.model": {Resources: go3mf.Resources{Objects: []*go3mf.Object{validMesh}}}},
			Resources: go3mf.Resources{Objects: []*go3mf.Object{
				{ID: 5, AnyAttr: go3mf.AnyAttr{&ObjectAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d481"}}, Components: []*go3mf.Component{
					{ObjectID: 1, AnyAttr: go3mf.AnyAttr{
						&ComponentAttr{Path: "/other.model", UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d480"},
					}}}}}}, Build: go3mf.Build{
				AnyAttr: go3mf.AnyAttr{&BuildAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d479"}}, Items: []*go3mf.Item{
					{ObjectID: 1, AnyAttr: go3mf.AnyAttr{&ItemAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d478", Path: "/other.model"}}},
				}}}, []error{
			fmt.Errorf("/other.model@Resources@Object#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
		}},
		{"items", &go3mf.Model{Build: go3mf.Build{
			AnyAttr: go3mf.AnyAttr{&BuildAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d479"}}, Items: []*go3mf.Item{
				{ObjectID: 1, AnyAttr: go3mf.AnyAttr{&ItemAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d478", Path: "/other.model"}}},
				{ObjectID: 1},
				{ObjectID: 1, AnyAttr: go3mf.AnyAttr{&ItemAttr{}}},
				{ObjectID: 1, AnyAttr: go3mf.AnyAttr{&ItemAttr{UUID: "a-b-c-d"}}},
			}},
			Childs:    map[string]*go3mf.ChildModel{"/other.model": {Resources: go3mf.Resources{Objects: []*go3mf.Object{validMesh}}}},
			Resources: go3mf.Resources{Objects: []*go3mf.Object{{ID: 1, Mesh: validMesh.Mesh}}}}, []error{
			fmt.Errorf("Build@Item#1: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("Build@Item#2: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("Build@Item#3: %v", ErrUUID),
			fmt.Errorf("/other.model@Resources@Object#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("Resources@Object#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
		}},
		{"components", &go3mf.Model{Resources: go3mf.Resources{
			Objects: []*go3mf.Object{
				{ID: 2, Mesh: validMesh.Mesh, AnyAttr: go3mf.AnyAttr{&ObjectAttr{UUID: "a-b-c-d"}}},
				{ID: 3, AnyAttr: go3mf.AnyAttr{&ObjectAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d483"}}, Components: []*go3mf.Component{
					{ObjectID: 2, AnyAttr: go3mf.AnyAttr{&ComponentAttr{}}},
					{ObjectID: 2, AnyAttr: go3mf.AnyAttr{&ComponentAttr{UUID: "a-b-c-d"}}},
					{ObjectID: 2},
				}},
			},
		}, Build: go3mf.Build{AnyAttr: go3mf.AnyAttr{&BuildAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d479"}}}}, []error{
			fmt.Errorf("Resources@Object#0: %v", ErrUUID),
			fmt.Errorf("Resources@Object#1@Component#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("Resources@Object#1@Component#1: %v", ErrUUID),
			fmt.Errorf("Resources@Object#1@Component#2: %v", &errors.MissingFieldError{Name: attrProdUUID}),
		}},
		{"child", &go3mf.Model{Build: go3mf.Build{AnyAttr: go3mf.AnyAttr{&BuildAttr{UUID: "f47ac10b-58cc-0372-8567-0e02b2c3d479"}}},
			Childs: map[string]*go3mf.ChildModel{
				"/b.model": {Resources: go3mf.Resources{Objects: []*go3mf.Object{validMesh}}},
				"/other.model": {Resources: go3mf.Resources{Objects: []*go3mf.Object{
					{ID: 2, Components: []*go3mf.Component{
						{ObjectID: 1, AnyAttr: go3mf.AnyAttr{&ComponentAttr{Path: "/b.model"}}},
					}},
				}}}}}, []error{
			fmt.Errorf("/b.model@Resources@Object#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("/other.model@Resources@Object#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("/other.model@Resources@Object#0@Component#0: %v", &errors.MissingFieldError{Name: attrProdUUID}),
			fmt.Errorf("/other.model@Resources@Object#0@Component#0: %v", ErrProdRefInNonRoot),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.model.Specs) == 0 {
				tt.model.WithSpec(&Spec{})
			}
			got := tt.model.Validate()
			if diff := deep.Equal(got.(*errors.List).Errors, tt.want); diff != nil {
				t.Errorf("Validate() = %v", diff)
			}
		})
	}
}
