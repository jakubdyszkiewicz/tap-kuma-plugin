package main

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_common_tap "github.com/envoyproxy/go-control-plane/envoy/extensions/common/tap/v3"
	envoy_tap "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/tap/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/hashicorp/go-hclog"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/externalpolicy"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"github.com/kumahq/kuma/pkg/xds/hooks"
	"google.golang.org/protobuf/proto"

	"jakubdyszkiewicz.io/tap-kuma-plugin/api"
)

type TapPolicyPlugin struct {
	logger hclog.Logger
}

var _ externalpolicy.ExternalPolicyPlugin = &TapPolicyPlugin{}

func (t TapPolicyPlugin) Descriptor() (model.ResourceTypeDescriptor, error) {
	descriptor := api.TapResourceTypeDescriptor
	descriptor.Resource = nil // we need to clear this, because gob will fail if it tries to send TapResource
	descriptor.ResourceList = nil
	return descriptor, nil
}

func (t TapPolicyPlugin) Filters() (hooks.Filters, error) {
	// User would provide filters what resources do they need
	// For example: "I need inbound listeners and Tap policy"
	// To save CPU, ee don't want to send all XDS and policies do the plugin.
	return hooks.Filters{}, nil
}

func (t TapPolicyPlugin) Modifications(data hooks.XDSHookData) (hooks.XDSHookModifications, error) {
	mods := hooks.XDSHookModifications{}

	// unmarshal data
	taps := api.TapResourceList{}
	if err := unmarshalExtensionResourceList(&taps, data.Policies); err != nil {
		return hooks.XDSHookModifications{}, err
	}

	dp := mesh.NewDataplaneResource()
	if err := unmarshalCoreResource(dp, data.Dataplane); err != nil {
		return hooks.XDSHookModifications{}, err
	}

	// match the policies
	var dataplanePolicies []policy.DataplanePolicy
	for _, tap := range taps.Items {
		dataplanePolicies = append(dataplanePolicies, tap)
	}

	matchedPolicy := policy.SelectDataplanePolicy(dp, dataplanePolicies)
	if matchedPolicy == nil { // no policy matched
		return hooks.XDSHookModifications{}, nil
	}

	// modify xds config
	for _, res := range data.Resources {
		if res.ID.Type != "Listener" {
			continue
		}
		if res.Origin != "inbound" {
			continue
		}
		l := envoy_listener.Listener{}
		if err := proto.Unmarshal(res.Resource, &l); err != nil {
			return hooks.XDSHookModifications{}, err
		}

		err := envoy_listeners.UpdateHTTPConnectionManager(l.FilterChains[0], func(manager *envoy_hcm.HttpConnectionManager) error {
			tap := envoy_tap.Tap{
				CommonConfig: &envoy_common_tap.CommonExtensionConfig{
					ConfigType: &envoy_common_tap.CommonExtensionConfig_AdminConfig{
						AdminConfig: &envoy_common_tap.AdminConfig{
							ConfigId: matchedPolicy.(*api.TapResource).Spec.Conf.Id,
						},
					},
				},
			}

			pbst, err := util_proto.MarshalAnyDeterministic(&tap)
			if err != nil {
				return err
			}

			filter := &envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.tap",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			}
			manager.HttpFilters = append([]*envoy_hcm.HttpFilter{filter}, manager.HttpFilters...)
			return nil
		})
		if err != nil {
			return hooks.XDSHookModifications{}, err
		}

		resBytes, err := proto.Marshal(&l)
		if err != nil {
			return hooks.XDSHookModifications{}, err
		}
		mods.Update = append(mods.Update, hooks.XDSResource{
			ID:       res.ID,
			Origin:   res.Origin,
			Resource: resBytes,
		})
	}
	return mods, nil
}
