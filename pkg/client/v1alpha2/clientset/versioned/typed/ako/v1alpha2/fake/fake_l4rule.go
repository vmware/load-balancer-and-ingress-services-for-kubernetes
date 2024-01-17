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

	v1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeL4Rules implements L4RuleInterface
type FakeL4Rules struct {
	Fake *FakeAkoV1alpha2
	ns   string
}

var l4rulesResource = v1alpha2.SchemeGroupVersion.WithResource("l4rules")

var l4rulesKind = v1alpha2.SchemeGroupVersion.WithKind("L4Rule")

// Get takes name of the l4Rule, and returns the corresponding l4Rule object, and an error if there is any.
func (c *FakeL4Rules) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.L4Rule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(l4rulesResource, c.ns, name), &v1alpha2.L4Rule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.L4Rule), err
}

// List takes label and field selectors, and returns the list of L4Rules that match those selectors.
func (c *FakeL4Rules) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.L4RuleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(l4rulesResource, l4rulesKind, c.ns, opts), &v1alpha2.L4RuleList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.L4RuleList{ListMeta: obj.(*v1alpha2.L4RuleList).ListMeta}
	for _, item := range obj.(*v1alpha2.L4RuleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested l4Rules.
func (c *FakeL4Rules) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(l4rulesResource, c.ns, opts))

}

// Create takes the representation of a l4Rule and creates it.  Returns the server's representation of the l4Rule, and an error, if there is any.
func (c *FakeL4Rules) Create(ctx context.Context, l4Rule *v1alpha2.L4Rule, opts v1.CreateOptions) (result *v1alpha2.L4Rule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(l4rulesResource, c.ns, l4Rule), &v1alpha2.L4Rule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.L4Rule), err
}

// Update takes the representation of a l4Rule and updates it. Returns the server's representation of the l4Rule, and an error, if there is any.
func (c *FakeL4Rules) Update(ctx context.Context, l4Rule *v1alpha2.L4Rule, opts v1.UpdateOptions) (result *v1alpha2.L4Rule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(l4rulesResource, c.ns, l4Rule), &v1alpha2.L4Rule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.L4Rule), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeL4Rules) UpdateStatus(ctx context.Context, l4Rule *v1alpha2.L4Rule, opts v1.UpdateOptions) (*v1alpha2.L4Rule, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(l4rulesResource, "status", c.ns, l4Rule), &v1alpha2.L4Rule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.L4Rule), err
}

// Delete takes name of the l4Rule and deletes it. Returns an error if one occurs.
func (c *FakeL4Rules) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(l4rulesResource, c.ns, name, opts), &v1alpha2.L4Rule{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeL4Rules) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(l4rulesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.L4RuleList{})
	return err
}

// Patch applies the patch and returns the patched l4Rule.
func (c *FakeL4Rules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.L4Rule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(l4rulesResource, c.ns, name, pt, data, subresources...), &v1alpha2.L4Rule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.L4Rule), err
}
