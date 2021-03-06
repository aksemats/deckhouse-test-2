/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: node-manager :: hooks :: discover_apiserver_endpoints ::", func() {
	const (
		stateSingleAddress = `
---
apiVersion: v1
kind: Endpoints
metadata:
  name: kubernetes
  namespace: default
subsets:
- addresses:
  - ip: 10.0.3.192
  ports:
  - name: https
    port: 6443
    protocol: TCP
`

		stateMultipleAddresses = `
---
apiVersion: v1
kind: Endpoints
metadata:
  name: kubernetes
  namespace: default
subsets:
- addresses:
  - ip: 10.0.3.192
  - ip: 10.0.3.193
  - ip: 10.0.3.194
  ports:
  - name: https
    port: 6443
    protocol: TCP
`

		stateMultupleAddressesWithDifferentPorts = `
---
apiVersion: v1
kind: Endpoints
metadata:
  name: kubernetes
  namespace: default
subsets:
- addresses:
  - ip: 10.0.3.192
  - ip: 10.0.3.193
  ports:
  - name: https
    port: 6443
    protocol: TCP
- addresses:
  - ip: 10.0.3.194
  ports:
  - name: https
    port: 6444
    protocol: TCP
`

		stateDeckhouseAPIServerPod = `
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver-0
  namespace: kube-system
  labels:
    component: kube-apiserver
    tier: control-plane
status:
  podIP: 192.168.199.233
  conditions:
  - status: "True"
    type: Ready
`
		stateDeckhouseAPIServerSecondPod = `
---
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver-1
  namespace: kube-system
  labels:
    component: kube-apiserver
    tier: control-plane
status:
  podIP: 192.168.199.244
  conditions:
  - status: "True"
    type: Ready
`
	)

	f := HookExecutionConfigInit(`{"nodeManager":{"internal": {}}}`, `{}`)

	Context("Endpoint default/kubernetes has single address in .subsets[]", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(stateSingleAddress, 1))
			f.RunHook()
		})

		It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443']", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443"]`))
		})

		Context("Someone added additional addresses to .subsets[]", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(stateMultipleAddresses, 1))
				f.RunHook()
			})

			It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443','10.0.3.193:6443','10.0.3.194:6443']", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443","10.0.3.193:6443","10.0.3.194:6443"]`))
			})

			Context("Someone added address with different port", func() {
				BeforeEach(func() {
					f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(stateMultupleAddressesWithDifferentPorts, 1))
					f.RunHook()
				})

				It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443','10.0.3.193:6443','10.0.3.194:6444']", func() {
					Expect(f).To(ExecuteSuccessfully())
					Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443","10.0.3.193:6443","10.0.3.194:6444"]`))
				})

				Context("Kube-apiserver pod is present", func() {
					BeforeEach(func() {
						f.BindingContexts.Set(f.KubeStateSet(stateMultupleAddressesWithDifferentPorts + stateDeckhouseAPIServerPod))
						f.RunHook()
					})

					It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443','10.0.3.193:6443','10.0.3.194:6444','192.168.199.233:6443']", func() {
						Expect(f).To(ExecuteSuccessfully())
						Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443","10.0.3.193:6443","10.0.3.194:6444","192.168.199.233:6443"]`))
					})

					Context("Second kube-apiserver pod is present", func() {
						BeforeEach(func() {
							f.BindingContexts.Set(f.KubeStateSet(stateMultupleAddressesWithDifferentPorts + stateDeckhouseAPIServerPod + stateDeckhouseAPIServerSecondPod))
							f.RunHook()
						})

						It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443','10.0.3.193:6443','10.0.3.194:6444','192.168.199.233:6443','192.168.199.244:6443']", func() {
							Expect(f).To(ExecuteSuccessfully())
							Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443","10.0.3.193:6443","10.0.3.194:6444","192.168.199.233:6443","192.168.199.244:6443"]`))
						})
					})

				})
			})
		})
	})

	Context("Endpoint default/kubernetes has multiple addresses in .subsets[]", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(stateMultipleAddresses, 1))
			f.RunHook()
		})

		It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443','10.0.3.193:6443','10.0.3.194:6443']", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443","10.0.3.193:6443","10.0.3.194:6443"]`))
		})

		Context("Someone set number of addresses in .subsets[] to one", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.KubeStateSetAndWaitForBindingContexts(stateSingleAddress, 1))
				f.RunHook()
			})

			It("`nodeManager.internal.clusterMasterAddresses` must be ['10.0.3.192:6443']", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(f.ValuesGet("nodeManager.internal.clusterMasterAddresses").String()).To(MatchJSON(`["10.0.3.192:6443"]`))
			})
		})
	})
})
