// MIT License
//
// Copyright (c) 2018 Jelmer Snoeck
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/jelmersnoeck/ingress-monitor/apis/ingressmonitor/v1alpha1"
	scheme "github.com/jelmersnoeck/ingress-monitor/pkg/client/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// IngressMonitorsGetter has a method to return a IngressMonitorInterface.
// A group's client should implement this interface.
type IngressMonitorsGetter interface {
	IngressMonitors(namespace string) IngressMonitorInterface
}

// IngressMonitorInterface has methods to work with IngressMonitor resources.
type IngressMonitorInterface interface {
	Create(*v1alpha1.IngressMonitor) (*v1alpha1.IngressMonitor, error)
	Update(*v1alpha1.IngressMonitor) (*v1alpha1.IngressMonitor, error)
	UpdateStatus(*v1alpha1.IngressMonitor) (*v1alpha1.IngressMonitor, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.IngressMonitor, error)
	List(opts v1.ListOptions) (*v1alpha1.IngressMonitorList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.IngressMonitor, err error)
	IngressMonitorExpansion
}

// ingressMonitors implements IngressMonitorInterface
type ingressMonitors struct {
	client rest.Interface
	ns     string
}

// newIngressMonitors returns a IngressMonitors
func newIngressMonitors(c *IngressmonitorV1alpha1Client, namespace string) *ingressMonitors {
	return &ingressMonitors{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the ingressMonitor, and returns the corresponding ingressMonitor object, and an error if there is any.
func (c *ingressMonitors) Get(name string, options v1.GetOptions) (result *v1alpha1.IngressMonitor, err error) {
	result = &v1alpha1.IngressMonitor{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ingressmonitors").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of IngressMonitors that match those selectors.
func (c *ingressMonitors) List(opts v1.ListOptions) (result *v1alpha1.IngressMonitorList, err error) {
	result = &v1alpha1.IngressMonitorList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ingressmonitors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested ingressMonitors.
func (c *ingressMonitors) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("ingressmonitors").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a ingressMonitor and creates it.  Returns the server's representation of the ingressMonitor, and an error, if there is any.
func (c *ingressMonitors) Create(ingressMonitor *v1alpha1.IngressMonitor) (result *v1alpha1.IngressMonitor, err error) {
	result = &v1alpha1.IngressMonitor{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("ingressmonitors").
		Body(ingressMonitor).
		Do().
		Into(result)
	return
}

// Update takes the representation of a ingressMonitor and updates it. Returns the server's representation of the ingressMonitor, and an error, if there is any.
func (c *ingressMonitors) Update(ingressMonitor *v1alpha1.IngressMonitor) (result *v1alpha1.IngressMonitor, err error) {
	result = &v1alpha1.IngressMonitor{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ingressmonitors").
		Name(ingressMonitor.Name).
		Body(ingressMonitor).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *ingressMonitors) UpdateStatus(ingressMonitor *v1alpha1.IngressMonitor) (result *v1alpha1.IngressMonitor, err error) {
	result = &v1alpha1.IngressMonitor{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ingressmonitors").
		Name(ingressMonitor.Name).
		SubResource("status").
		Body(ingressMonitor).
		Do().
		Into(result)
	return
}

// Delete takes name of the ingressMonitor and deletes it. Returns an error if one occurs.
func (c *ingressMonitors) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ingressmonitors").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *ingressMonitors) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ingressmonitors").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched ingressMonitor.
func (c *ingressMonitors) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.IngressMonitor, err error) {
	result = &v1alpha1.IngressMonitor{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("ingressmonitors").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
