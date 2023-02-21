package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	client dynamic.Interface
}

func NewClient(client dynamic.Interface) *Client {
	return &Client{client: client}
}

func (s *Client) Get(ctx context.Context, namespace, name string) (*Subscription, error) {
	object, err := s.client.Resource(groupVersionResource()).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscription(object)
}

func (s *Client) Create(ctx context.Context, subscription *Subscription) (*Subscription, error) {
	object, err := toUnstructured(subscription)
	if err != nil {
		return nil, err
	}
	object, err = s.client.Resource(groupVersionResource()).Namespace(subscription.Namespace).Create(ctx, object, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscription(object)
}

func (s *Client) Update(ctx context.Context, subscription *Subscription) (*Subscription, error) {
	object, err := toUnstructured(subscription)
	if err != nil {
		return nil, err
	}
	object, err = s.client.Resource(groupVersionResource()).Namespace(subscription.Namespace).Update(ctx, object, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscription(object)
}

func (s *Client) Delete(ctx context.Context, namespace, name string) error {
	return s.client.Resource(groupVersionResource()).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func toUnstructured(subscription *Subscription) (*unstructured.Unstructured, error) {
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&subscription)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func toSubscription(object *unstructured.Unstructured) (*Subscription, error) {
	subscription := &Subscription{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func groupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  GroupVersion.Version,
		Group:    GroupVersion.Group,
		Resource: Resource,
	}
}
