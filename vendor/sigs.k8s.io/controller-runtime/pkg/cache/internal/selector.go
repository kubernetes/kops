package internal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SelectorsByGVK associate a GroupVersionKind to a field/label selector
type SelectorsByGVK map[schema.GroupVersionKind]Selector

// Selector specify the label/field selector to fill in ListOptions
type Selector struct {
	Label labels.Selector
	Field fields.Selector
}

// ApplyToList fill in ListOptions LabelSelector and FieldSelector if needed
func (s Selector) ApplyToList(listOpts *metav1.ListOptions) {
	if s.Label != nil {
		listOpts.LabelSelector = s.Label.String()
	}
	if s.Field != nil {
		listOpts.FieldSelector = s.Field.String()
	}
}
