package production

import (
	"testing"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
)

var _ encoding.Decoder = new(Spec)
var _ encoding.PostProcessorDecoder = new(Spec)
var _ encoding.PreProcessEncoder = new(Spec)
var _ go3mf.Spec = new(Spec)
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
