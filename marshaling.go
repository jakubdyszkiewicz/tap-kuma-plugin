package main

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
	model2 "github.com/kumahq/kuma/pkg/test/resources/model"
	proto2 "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/hooks"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// TODO: move to core Kuma

func unmarshalExtensionResourceList(list model.ResourceList, policies []hooks.Policy) error {
	for _, policy := range policies {
		if policy.Type != string(list.GetItemType()) {
			continue
		}
		st := structpb.Struct{}
		if err := proto.Unmarshal(policy.Spec, &st); err != nil {
			return err
		}

		res := list.NewItem()
		tapBytes, err := st.MarshalJSON()
		if err != nil {
			return err
		}
		if err := proto2.FromJSON(tapBytes, res.GetSpec()); err != nil {
			return err
		}

		res.SetMeta(&model2.ResourceMeta{
			Mesh: policy.Mesh,
			Name: policy.Name,
		})
		if err := list.AddItem(res); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalCoreResource(resource model.Resource, policy hooks.Policy) error {
	if err := proto.Unmarshal(policy.Spec, resource.GetSpec()); err != nil {
		return err
	}
	resource.SetMeta(&model2.ResourceMeta{
		Mesh: policy.Mesh,
		Name: policy.Name,
	})
	return nil
}
