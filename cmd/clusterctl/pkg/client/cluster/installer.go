/*
Copyright 2019 The Kubernetes Authors.

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

package cluster

import (
	"sigs.k8s.io/cluster-api/cmd/clusterctl/pkg/client/repository"
)

// ProviderInstaller defines methods for enforcing consistency rules for provider installation.
type ProviderInstaller interface {
	// Add adds a provider to the install queue.
	// NB. By deferring the installation, the installer service can perform validation of the target state of the management cluster
	// before actually starting the installation of new providers.
	Add(repository.Components) error

	// Install performs the installation of the providers ready in the install queue.
	Install() ([]repository.Components, error)
}

// providerInstaller implements ProviderInstaller
type providerInstaller struct {
	proxy              Proxy
	providerComponents ComponentsClient
	providerInventory  InventoryClient
	installQueue       []repository.Components
}

var _ ProviderInstaller = &providerInstaller{}

func (i *providerInstaller) Add(components repository.Components) error {
	if err := i.providerInventory.Validate(components.InventoryObject()); err != nil {
		return err
	}

	i.installQueue = append(i.installQueue, components)
	return nil
}

func (i *providerInstaller) Install() ([]repository.Components, error) {
	ret := make([]repository.Components, 0, len(i.installQueue))
	for _, components := range i.installQueue {
		if err := installComponentsAndUpdateInventory(components, i.providerComponents, i.providerInventory); err != nil {
			return nil, err
		}

		ret = append(ret, components)
	}
	return ret, nil
}

func installComponentsAndUpdateInventory(components repository.Components, providerComponents ComponentsClient, providerInventory InventoryClient) error {
	if err := providerComponents.Create(components); err != nil {
		return err
	}

	if err := providerInventory.Create(components.InventoryObject()); err != nil {
		return err
	}

	return nil
}

func newProviderInstaller(proxy Proxy, providerMetadata InventoryClient, providerComponents ComponentsClient) *providerInstaller {
	return &providerInstaller{
		proxy:              proxy,
		providerInventory:  providerMetadata,
		providerComponents: providerComponents,
	}
}
