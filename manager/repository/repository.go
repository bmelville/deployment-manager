/*
Copyright 2015 The Kubernetes Authors All rights reserved.
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

// Package repository implements a deployment repository using a map.
// It can be easily replaced by a deployment repository that uses some
// form of persistent storage.
package repository

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kubernetes/deployment-manager/common"
)

// Repository manages storage for all Deployment Manager entities, as well as
// the common operations to store, access and manage them.
type Repository interface {
	// Deployments.
	ListDeployments() ([]common.Deployment, error)
	GetDeployment(name string) (*common.Deployment, error)
	GetValidDeployment(name string) (*common.Deployment, error)
	CreateDeployment(name string) (*common.Deployment, error)
	DeleteDeployment(name string, forget bool) (*common.Deployment, error)
	SetDeploymentState(name string, state *common.DeploymentState) error

	// Manifests.
	AddManifest(deploymentName string, manifest *common.Manifest) error
	SetManifest(deploymentName string, manifest *common.Manifest) error
	ListManifests(deploymentName string) (map[string]*common.Manifest, error)
	GetManifest(deploymentName string, manifestName string) (*common.Manifest, error)
	GetLatestManifest(deploymentName string) (*common.Manifest, error)

	// Types.
	ListTypes() []string
	GetTypeInstances(typeName string) []*common.TypeInstance
	ClearTypeInstances(deploymentName string)
	SetTypeInstances(deploymentName string, instances map[string][]*common.TypeInstance)
}

// deploymentTypeInstanceMap stores type instances mapped by deployment name.
// This allows for simple updating and deleting of per-deployment instances
// when deployments are created/updated/deleted.
type deploymentTypeInstanceMap map[string][]*common.TypeInstance
type typeInstanceMap map[string]deploymentTypeInstanceMap

type mapBasedRepository struct {
	sync.RWMutex
	deployments map[string]common.Deployment
	manifests   map[string]map[string]*common.Manifest
	instances   typeInstanceMap
}

// NewMapBasedRepository returns a new map based repository.
func NewMapBasedRepository() Repository {
	return &mapBasedRepository{
		deployments: make(map[string]common.Deployment, 0),
		manifests:   make(map[string]map[string]*common.Manifest, 0),
		instances:   typeInstanceMap{},
	}
}

// ListDeployments returns of all of the deployments in the repository.
func (r *mapBasedRepository) ListDeployments() ([]common.Deployment, error) {
	r.RLock()
	defer r.RUnlock()

	l := []common.Deployment{}
	for _, deployment := range r.deployments {
		l = append(l, deployment)
	}

	return l, nil
}

// GetDeployment returns the deployment with the supplied name.
// If the deployment is not found, it returns an error.
func (r *mapBasedRepository) GetDeployment(name string) (*common.Deployment, error) {
	d, ok := r.deployments[name]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", name)
	}
	return &d, nil
}

// GetValidDeployment returns the deployment with the supplied name.
// If the deployment is not found or marked as deleted, it returns an error.
func (r *mapBasedRepository) GetValidDeployment(name string) (*common.Deployment, error) {
	d, err := r.GetDeployment(name)
	if err != nil {
		return nil, err
	}

	if d.State.Status == common.DeletedStatus {
		return nil, fmt.Errorf("deployment %s is deleted", name)
	}

	return d, nil
}

// SetDeploymentState sets the DeploymentState of the deployment and updates ModifiedAt
func (r *mapBasedRepository) SetDeploymentState(name string, state *common.DeploymentState) error {
	return func() error {
		r.Lock()
		defer r.Unlock()

		d, err := r.GetValidDeployment(name)
		if err != nil {
			return err
		}

		d.State = state
		d.ModifiedAt = time.Now()
		r.deployments[name] = *d
		return nil
	}()
}

// CreateDeployment creates a new deployment and stores it in the repository.
func (r *mapBasedRepository) CreateDeployment(name string) (*common.Deployment, error) {
	d, err := func() (*common.Deployment, error) {
		r.Lock()
		defer r.Unlock()

		exists, _ := r.GetValidDeployment(name)
		if exists != nil {
			return nil, fmt.Errorf("Deployment %s already exists", name)
		}

		d := common.NewDeployment(name)
		d.DeployedAt = time.Now()
		r.deployments[name] = *d
		return d, nil
	}()

	if err != nil {
		return nil, err
	}

	log.Printf("created deployment: %v", d)
	return d, nil
}

// AddManifest adds a manifest to the repository and repoints the latest
// manifest to it for the corresponding deployment.
func (r *mapBasedRepository) AddManifest(deploymentName string, manifest *common.Manifest) error {
	r.Lock()
	defer r.Unlock()

	l, err := r.listManifestsForDeployment(deploymentName)
	if err != nil {
		return err
	}

	// Make sure the manifest doesn't already exist, and if not, add the manifest to
	// map of manifests this deployment has
	if _, ok := l[manifest.Name]; ok {
		return fmt.Errorf("Manifest %s already exists in deployment %s", manifest.Name, deploymentName)
	}

	d, err := r.GetValidDeployment(deploymentName)
	if err != nil {
		return err
	}

	l[manifest.Name] = manifest
	d.LatestManifest = manifest.Name
	r.deployments[deploymentName] = *d

	log.Printf("Added manifest %s to deployment: %s", manifest.Name, deploymentName)
	return nil
}

// SetManifest sets an existing manifest in the repository to provided
// manifest.
func (r *mapBasedRepository) SetManifest(deploymentName string, manifest *common.Manifest) error {
	r.Lock()
	defer r.Unlock()

	l, err := r.listManifestsForDeployment(deploymentName)
	if err != nil {
		return err
	}

	l[manifest.Name] = manifest
	return nil
}

// DeleteDeployment deletes the deployment with the supplied name.
// If forget is true, then the deployment is removed from the repository.
// Otherwise, it is marked as deleted and retained.
func (r *mapBasedRepository) DeleteDeployment(name string, forget bool) (*common.Deployment, error) {
	d, err := func() (*common.Deployment, error) {
		r.Lock()
		defer r.Unlock()

		d, err := r.GetValidDeployment(name)
		if err != nil {
			return nil, err
		}

		if !forget {
			d.DeletedAt = time.Now()
			d.State = &common.DeploymentState{Status: common.DeletedStatus}
			r.deployments[name] = *d
		} else {
			delete(r.deployments, name)
			delete(r.manifests, name)
			d.LatestManifest = ""
		}

		return d, nil
	}()

	if err != nil {
		return nil, err
	}

	log.Printf("deleted deployment: %v", d)
	return d, nil
}

func (r *mapBasedRepository) ListManifests(deploymentName string) (map[string]*common.Manifest, error) {
	r.Lock()
	defer r.Unlock()

	_, err := r.GetValidDeployment(deploymentName)
	if err != nil {
		return nil, err
	}

	return r.listManifestsForDeployment(deploymentName)
}

func (r *mapBasedRepository) listManifestsForDeployment(deploymentName string) (map[string]*common.Manifest, error) {
	l, ok := r.manifests[deploymentName]
	if !ok {
		l = make(map[string]*common.Manifest, 0)
		r.manifests[deploymentName] = l
	}

	return l, nil
}

func (r *mapBasedRepository) GetManifest(deploymentName string, manifestName string) (*common.Manifest, error) {
	r.Lock()
	defer r.Unlock()

	_, err := r.GetValidDeployment(deploymentName)
	if err != nil {
		return nil, err
	}

	return r.getManifestForDeployment(deploymentName, manifestName)
}

func (r *mapBasedRepository) getManifestForDeployment(deploymentName string, manifestName string) (*common.Manifest, error) {
	l, err := r.listManifestsForDeployment(deploymentName)
	if err != nil {
		return nil, err
	}

	m, ok := l[manifestName]
	if !ok {
		return nil, fmt.Errorf("manifest %s not found in deployment %s", manifestName, deploymentName)
	}

	return m, nil
}

// GetLatestManifest returns the latest manifest for a given deployment,
// which by definition is the manifest with the largest time stamp.
func (r *mapBasedRepository) GetLatestManifest(deploymentName string) (*common.Manifest, error) {
	r.Lock()
	defer r.Unlock()

	d, err := r.GetValidDeployment(deploymentName)
	if err != nil {
		return nil, err
	}

	return r.getManifestForDeployment(deploymentName, d.LatestManifest)
}

// ListTypes returns all types known from existing instances.
func (r *mapBasedRepository) ListTypes() []string {
	var keys []string
	for k := range r.instances {
		keys = append(keys, k)
	}

	return keys
}

// GetTypeInstances returns all instances of a given type. If type is empty,
// returns all instances for all types.
func (r *mapBasedRepository) GetTypeInstances(typeName string) []*common.TypeInstance {
	r.Lock()
	defer r.Unlock()

	var instances []*common.TypeInstance
	for t, dInstMap := range r.instances {
		if t == typeName || typeName == "all" {
			for _, i := range dInstMap {
				instances = append(instances, i...)
			}
		}
	}

	return instances
}

// ClearTypeInstances deletes all instances associated with the given
// deployment name from the type instance repository.
func (r *mapBasedRepository) ClearTypeInstances(deploymentName string) {
	r.Lock()
	defer r.Unlock()

	for t, dMap := range r.instances {
		delete(dMap, deploymentName)
		if len(dMap) == 0 {
			delete(r.instances, t)
		}
	}
}

// SetTypeInstances sets all type instances for a given deployment name.
//
// To clear the current set of instances first, caller should first use
// ClearTypeInstances().
func (r *mapBasedRepository) SetTypeInstances(deploymentName string, instances map[string][]*common.TypeInstance) {
	r.Lock()
	defer r.Unlock()

	// Add each instance list to the appropriate type map.
	for t, is := range instances {
		if r.instances[t] == nil {
			r.instances[t] = make(deploymentTypeInstanceMap)
		}

		r.instances[t][deploymentName] = is
	}
}
