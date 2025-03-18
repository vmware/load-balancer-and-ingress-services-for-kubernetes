/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeHealthMonitors implements HealthMonitorInterface
type FakeHealthMonitors struct {
	Fake *FakeAkoV1alpha1
	ns   string
}

var healthmonitorsResource = v1alpha1.SchemeGroupVersion.WithResource("healthmonitors")

var healthmonitorsKind = v1alpha1.SchemeGroupVersion.WithKind("HealthMonitor")

// Get takes name of the healthMonitor, and returns the corresponding healthMonitor object, and an error if there is any.
func (c *FakeHealthMonitors) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.HealthMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(healthmonitorsResource, c.ns, name), &v1alpha1.HealthMonitor{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HealthMonitor), err
}

// List takes label and field selectors, and returns the list of HealthMonitors that match those selectors.
func (c *FakeHealthMonitors) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.HealthMonitorList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(healthmonitorsResource, healthmonitorsKind, c.ns, opts), &v1alpha1.HealthMonitorList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.HealthMonitorList{ListMeta: obj.(*v1alpha1.HealthMonitorList).ListMeta}
	for _, item := range obj.(*v1alpha1.HealthMonitorList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested healthMonitors.
func (c *FakeHealthMonitors) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(healthmonitorsResource, c.ns, opts))

}

// Create takes the representation of a healthMonitor and creates it.  Returns the server's representation of the healthMonitor, and an error, if there is any.
func (c *FakeHealthMonitors) Create(ctx context.Context, healthMonitor *v1alpha1.HealthMonitor, opts v1.CreateOptions) (result *v1alpha1.HealthMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(healthmonitorsResource, c.ns, healthMonitor), &v1alpha1.HealthMonitor{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HealthMonitor), err
}

// Update takes the representation of a healthMonitor and updates it. Returns the server's representation of the healthMonitor, and an error, if there is any.
func (c *FakeHealthMonitors) Update(ctx context.Context, healthMonitor *v1alpha1.HealthMonitor, opts v1.UpdateOptions) (result *v1alpha1.HealthMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(healthmonitorsResource, c.ns, healthMonitor), &v1alpha1.HealthMonitor{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HealthMonitor), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeHealthMonitors) UpdateStatus(ctx context.Context, healthMonitor *v1alpha1.HealthMonitor, opts v1.UpdateOptions) (*v1alpha1.HealthMonitor, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(healthmonitorsResource, "status", c.ns, healthMonitor), &v1alpha1.HealthMonitor{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HealthMonitor), err
}

// Delete takes name of the healthMonitor and deletes it. Returns an error if one occurs.
func (c *FakeHealthMonitors) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(healthmonitorsResource, c.ns, name, opts), &v1alpha1.HealthMonitor{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHealthMonitors) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(healthmonitorsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.HealthMonitorList{})
	return err
}

// Patch applies the patch and returns the patched healthMonitor.
func (c *FakeHealthMonitors) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.HealthMonitor, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(healthmonitorsResource, c.ns, name, pt, data, subresources...), &v1alpha1.HealthMonitor{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HealthMonitor), err
}
