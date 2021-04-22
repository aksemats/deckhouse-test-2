package hooks

import (
	"errors"
	"fmt"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube_events_manager/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/kube-proxy",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:              "kube_api_ep",
			ApiVersion:        "v1",
			Kind:              "Endpoints",
			NamespaceSelector: &types.NamespaceSelector{NameSelector: &types.NameSelector{MatchNames: []string{"default"}}},
			NameSelector:      &types.NameSelector{MatchNames: []string{"kubernetes"}},
			FilterFunc:        applyKubernetesAPIEndpointsFilter,
		},
	},
}, discoverAPIEndpointsHandler)

// KubernetesAPIEndpoints discovers kube api endpoints
type KubernetesAPIEndpoints struct {
	HostPort []string
}

func applyKubernetesAPIEndpointsFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	endpoint := &v1.Endpoints{}
	err := sdk.FromUnstructured(obj, endpoint)
	if err != nil {
		return nil, err
	}

	mh := &KubernetesAPIEndpoints{}

	for _, subset := range endpoint.Subsets {
		for _, address := range subset.Addresses {
			ip := address.IP
			for _, port := range subset.Ports {
				mh.HostPort = append(mh.HostPort, fmt.Sprintf("%s:%d", ip, port.Port))
			}
		}
	}

	return mh, nil
}

func discoverAPIEndpointsHandler(input *go_hook.HookInput) error {
	ep, ok := input.Snapshots["kube_api_ep"]
	if !ok {
		return errors.New("no endpoints snapshot")
	}

	if len(ep) == 0 {
		input.LogEntry.Error("kubernetes endpoints not found")
		return nil
	}

	fpp := ep[0].(*KubernetesAPIEndpoints)

	if len(fpp.HostPort) == 0 {
		return errors.New("no kubernetes apiserver endpoints host:port specified")
	}

	input.LogEntry.Infof("cluster master addresses: %v", fpp.HostPort)

	input.Values.Set("kubeProxy.internal.clusterMasterAddresses", fpp.HostPort)

	return nil
}