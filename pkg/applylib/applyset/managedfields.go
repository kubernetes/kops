/*
Copyright 2022 The Kubernetes Authors.

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

package applyset

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/handlers/fieldmanager"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

// ManagedFieldsMigrator manages the migration of field managers from client-side managers to the server-side manager.
type ManagedFieldsMigrator struct {
	Client     *UnstructuredClient
	NewManager string
}

// Migrate migrates from client-side field managers to the NewManager (with an Apply operation).
// This is needed to move from client-side apply to server-side apply.
func (m *ManagedFieldsMigrator) Migrate(ctx context.Context, obj *unstructured.Unstructured) error {
	managedFieldPatch, err := m.createManagedFieldPatch(obj)
	if err != nil {
		return fmt.Errorf("failed to create managed-fields patch: %w", err)
	}
	if managedFieldPatch != nil {
		gvk := obj.GroupVersionKind()
		nn := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
		_, err := m.Client.Patch(ctx, gvk, nn, types.MergePatchType, managedFieldPatch, metav1.PatchOptions{})
		if err != nil {
			return fmt.Errorf("failed to patch object managed-fields for %q: %w", obj.GetName(), err)
		}
	}
	return nil
}

// createManagedFieldPatch constructs a patch to combine managed fields.
// It returns nil if no patch is needed.
func (m *ManagedFieldsMigrator) createManagedFieldPatch(currentObject *unstructured.Unstructured) ([]byte, error) {
	if currentObject == nil {
		return nil, nil
	}
	needPatch := false
	fixedManagedFields := []metav1.ManagedFieldsEntry{}
	for _, managedField := range currentObject.GetManagedFields() {
		fixedManagedField := managedField.DeepCopy()
		if managedField.Manager == "kubectl-edit" || managedField.Manager == "kubectl-client-side-apply" {
			needPatch = true
			fixedManagedField.Manager = m.NewManager
			fixedManagedField.Operation = metav1.ManagedFieldsOperationApply
		}
		// In case we have an existing Update operation
		if fixedManagedField.Manager == m.NewManager && fixedManagedField.Operation == "Update" {
			needPatch = true
			fixedManagedField.Operation = metav1.ManagedFieldsOperationApply
		}
		fixedManagedFields = append(fixedManagedFields, *fixedManagedField)
	}
	if !needPatch {
		return nil, nil
	}

	merged, err := mergeFieldManagers(fixedManagedFields)
	if err != nil {
		return nil, err
	}
	fixedManagedFields = merged

	// Ensure patch is stable, mostly for tests
	sort.Slice(fixedManagedFields, func(i, j int) bool {
		if fixedManagedFields[i].Manager != fixedManagedFields[j].Manager {
			return fixedManagedFields[i].Manager < fixedManagedFields[j].Manager
		}
		if fixedManagedFields[i].Subresource != fixedManagedFields[j].Subresource {
			return fixedManagedFields[i].Subresource < fixedManagedFields[j].Subresource
		}
		if fixedManagedFields[i].Operation != fixedManagedFields[j].Operation {
			return fixedManagedFields[i].Operation < fixedManagedFields[j].Operation
		}
		return false
	})

	meta := &metav1.ObjectMeta{}
	meta.SetManagedFields(fixedManagedFields)
	patchObject := map[string]interface{}{
		"metadata": meta,
	}

	// MarshalIndent is a little less efficient, but makes this much more readable (also helps tests)
	jsonData, err := json.MarshalIndent(patchObject, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marsal %q into json: %w", currentObject.GetName(), err)
	}
	return jsonData, nil
}

// fieldManagerKey is the primary key for a ManagedFieldEntry
type fieldManagerKey struct {
	Manager     string
	Operation   metav1.ManagedFieldsOperationType
	Subresource string
}

// mergeFieldManagers merges the managed fields from identical field managers.
// If we don't do this, the apiserver will not currently construct the union for duplicate keys.
func mergeFieldManagers(managedFields []metav1.ManagedFieldsEntry) ([]metav1.ManagedFieldsEntry, error) {
	byKey := make(map[fieldManagerKey][]metav1.ManagedFieldsEntry)
	for _, f := range managedFields {
		k := fieldManagerKey{
			Manager:     f.Manager,
			Operation:   f.Operation,
			Subresource: f.Subresource,
		}

		byKey[k] = append(byKey[k], f)
	}

	var result []metav1.ManagedFieldsEntry
	for k := range byKey {
		managers := byKey[k]
		if len(managers) > 1 {
			fieldSet, err := mergeManagedFields(managers)
			if err != nil {
				return nil, err
			}
			encoded, err := fieldSet.ToJSON()
			if err != nil {
				return nil, err
			}
			managers[0].FieldsV1.Raw = encoded
		}
		result = append(result, managers[0])
	}
	return result, nil
}

// mergeManagedFields merges a set of ManagedFieldEntry managed fields, that are expected to have the same key.
func mergeManagedFields(managedFields []metav1.ManagedFieldsEntry) (*fieldpath.Set, error) {
	if len(managedFields) == 0 {
		return nil, fmt.Errorf("no managed fields supplied")
	}

	union, err := toFieldPathSet(&managedFields[0])
	if err != nil {
		return nil, err
	}

	for i := range managedFields {
		if i == 0 {
			continue
		}
		m := &managedFields[i]
		if managedFields[0].APIVersion != m.APIVersion {
			return nil, fmt.Errorf("cannot merge ManagedFieldsEntry apiVersion %q with apiVersion %q", managedFields[0].APIVersion, m.APIVersion)
		}

		set, err := toFieldPathSet(m)
		if err != nil {
			return nil, err
		}
		union = union.Union(set)
	}
	return union, nil
}

// toFieldPathSet converts an encoded ManagedFieldsEntry to a set of managed fields (a fieldpath.Set)
func toFieldPathSet(fields *metav1.ManagedFieldsEntry) (*fieldpath.Set, error) {
	decoded, err := fieldmanager.DecodeManagedFields([]metav1.ManagedFieldsEntry{*fields})
	if err != nil {
		return nil, err
	}
	if len(decoded.Fields()) != 1 {
		return nil, fmt.Errorf("expected a single managed fields entry, but got %d", len(decoded.Fields()))
	}
	for _, fieldSet := range decoded.Fields() {
		return fieldSet.Set(), nil
	}
	return nil, fmt.Errorf("no fields were decoded")
}
