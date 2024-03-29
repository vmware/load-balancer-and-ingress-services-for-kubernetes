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

package v1alpha2

import (
	"context"
	"time"

	v1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	scheme "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha2/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// L7RulesGetter has a method to return a L7RuleInterface.
// A group's client should implement this interface.
type L7RulesGetter interface {
	L7Rules(namespace string) L7RuleInterface
}

// L7RuleInterface has methods to work with L7Rule resources.
type L7RuleInterface interface {
	Create(ctx context.Context, l7Rule *v1alpha2.L7Rule, opts v1.CreateOptions) (*v1alpha2.L7Rule, error)
	Update(ctx context.Context, l7Rule *v1alpha2.L7Rule, opts v1.UpdateOptions) (*v1alpha2.L7Rule, error)
	UpdateStatus(ctx context.Context, l7Rule *v1alpha2.L7Rule, opts v1.UpdateOptions) (*v1alpha2.L7Rule, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.L7Rule, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.L7RuleList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.L7Rule, err error)
	L7RuleExpansion
}

// l7Rules implements L7RuleInterface
type l7Rules struct {
	client rest.Interface
	ns     string
}

// newL7Rules returns a L7Rules
func newL7Rules(c *AkoV1alpha2Client, namespace string) *l7Rules {
	return &l7Rules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the l7Rule, and returns the corresponding l7Rule object, and an error if there is any.
func (c *l7Rules) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.L7Rule, err error) {
	result = &v1alpha2.L7Rule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("l7rules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of L7Rules that match those selectors.
func (c *l7Rules) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.L7RuleList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.L7RuleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("l7rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested l7Rules.
func (c *l7Rules) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("l7rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a l7Rule and creates it.  Returns the server's representation of the l7Rule, and an error, if there is any.
func (c *l7Rules) Create(ctx context.Context, l7Rule *v1alpha2.L7Rule, opts v1.CreateOptions) (result *v1alpha2.L7Rule, err error) {
	result = &v1alpha2.L7Rule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("l7rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(l7Rule).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a l7Rule and updates it. Returns the server's representation of the l7Rule, and an error, if there is any.
func (c *l7Rules) Update(ctx context.Context, l7Rule *v1alpha2.L7Rule, opts v1.UpdateOptions) (result *v1alpha2.L7Rule, err error) {
	result = &v1alpha2.L7Rule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("l7rules").
		Name(l7Rule.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(l7Rule).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *l7Rules) UpdateStatus(ctx context.Context, l7Rule *v1alpha2.L7Rule, opts v1.UpdateOptions) (result *v1alpha2.L7Rule, err error) {
	result = &v1alpha2.L7Rule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("l7rules").
		Name(l7Rule.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(l7Rule).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the l7Rule and deletes it. Returns an error if one occurs.
func (c *l7Rules) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("l7rules").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *l7Rules) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("l7rules").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched l7Rule.
func (c *l7Rules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.L7Rule, err error) {
	result = &v1alpha2.L7Rule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("l7rules").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
