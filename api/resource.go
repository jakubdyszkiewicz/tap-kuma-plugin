package api

import (
	"fmt"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	TapType model.ResourceType = "Tap"
)

var _ model.Resource = &TapResource{}

type TapResource struct {
	Meta model.ResourceMeta
	Spec *Tap
}

func NewTapResource() *TapResource {
	return &TapResource{
		Spec: &Tap{},
	}
}

func (t *TapResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *TapResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *TapResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *TapResource) Selectors() []*v1alpha1.Selector {
	return t.Spec.GetSelectors()
}

func (t *TapResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*Tap)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &Tap{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *TapResource) Descriptor() model.ResourceTypeDescriptor {
	return TapResourceTypeDescriptor
}

var _ model.ResourceList = &TapResourceList{}

type TapResourceList struct {
	Items      []*TapResource
	Pagination model.Pagination
}

func (l *TapResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *TapResourceList) GetItemType() model.ResourceType {
	return TapType
}

func (l *TapResourceList) NewItem() model.Resource {
	return NewTapResource()
}

func (l *TapResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TapResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TapResource)(nil), r)
	}
}

func (l *TapResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var TapResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                TapType,
	Resource:            NewTapResource(),
	ResourceList:        &TapResourceList{},
	ReadOnly:            false,
	AdminOnly:           false,
	Scope:               model.ScopeMesh,
	KDSFlags:            model.FromGlobalToZone,
	WsPath:              "taps",
	KumactlArg:          "tap",
	KumactlListArg:      "taps",
	AllowToInspect:      true,
	IsPolicy:            true,
	SingularDisplayName: "Tap",
	PluralDisplayName:   "Taps",
	IsExperimental:      false,
}
