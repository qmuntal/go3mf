package production

import (
	"testing"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
)

var _ encoding.Decoder = new(Spec)
var _ spec.Spec = new(Spec)
var _ spec.ValidatorSpec = new(Spec)
var _ encoding.MarshalerAttr = new(BuildAttr)
var _ encoding.MarshalerAttr = new(ItemAttr)
var _ encoding.MarshalerAttr = new(ComponentAttr)
var _ encoding.MarshalerAttr = new(ObjectAttr)

func TestComponentAttr_ObjectPath(t *testing.T) {
	tests := []struct {
		name string
		p    *ComponentAttr
		want string
	}{
		{"empty", new(ComponentAttr), ""},
		{"path", &ComponentAttr{Path: "/a.model"}, "/a.model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ObjectPath(); got != tt.want {
				t.Errorf("ComponentAttr.ObjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemAttr_ObjectPath(t *testing.T) {
	tests := []struct {
		name string
		p    *ItemAttr
		want string
	}{
		{"empty", new(ItemAttr), ""},
		{"path", &ItemAttr{Path: "/a.model"}, "/a.model"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ObjectPath(); got != tt.want {
				t.Errorf("ItemAttr.ObjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpec_AddMissingUUIDs(t *testing.T) {
	components := &go3mf.Object{
		ID:         20,
		Components: []*go3mf.Component{{ObjectID: 8}},
	}
	m := &go3mf.Model{Path: "/3D/3dmodel.model", Build: go3mf.Build{}}
	m.Resources = go3mf.Resources{Objects: []*go3mf.Object{components}}
	m.Build.Items = append(m.Build.Items, &go3mf.Item{ObjectID: 20}, &go3mf.Item{ObjectID: 8})
	AddMissingUUIDs(m)
	if len(m.Build.AnyAttr) == 0 {
		t.Errorf("AddMissingUUIDs() should have filled build attrs")
	}
	if len(m.Build.Items[0].AnyAttr) == 0 {
		t.Errorf("AddMissingUUIDs() should have filled item attrs")
	}
	if len(m.Build.Items[1].AnyAttr) == 0 {
		t.Errorf("AddMissingUUIDs() should have filled item attrs")
	}
	if len(m.Resources.Objects[0].AnyAttr) == 0 {
		t.Errorf("AddMissingUUIDs() should have filled object attrs")
	}
	if len(m.Resources.Objects[0].Components[0].AnyAttr) == 0 {
		t.Errorf("AddMissingUUIDs() should have filled object attrs")
	}
}
