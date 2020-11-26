/*
Copyright 2019 The Jetstack cert-manager contributors.

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

package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/clock"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	logf "github.com/jetstack/cert-manager/pkg/logs"
)

// Clock is defined as a package var so it can be stubbed out during tests.
var Clock clock.Clock = clock.RealClock{}

// IssuerHasCondition will return true if the given GenericIssuer has a
// condition matching the provided IssuerCondition.
// Only the Type and Status field will be used in the comparison, meaning that
// this function will return 'true' even if the Reason, Message and
// LastTransitionTime fields do not match.
func IssuerHasCondition(i cmapi.GenericIssuer, c cmapi.IssuerCondition) bool {
	if i == nil {
		return false
	}
	existingConditions := i.GetStatus().Conditions
	for _, cond := range existingConditions {
		if c.Type == cond.Type && c.Status == cond.Status {
			return true
		}
	}
	return false
}

// SetIssuerCondition will set a 'condition' on the given GenericIssuer.
// - If no condition of the same type already exists, the condition will be
//   inserted with the LastTransitionTime set to the current time.
// - If a condition of the same type and state already exists, the condition
//   will be updated but the LastTransitionTime will not be modified.
// - If a condition of the same type and different state already exists, the
//   condition will be updated and the LastTransitionTime set to the current
//   time.
// This function works with both Issuer and ClusterIssuer resources.
func SetIssuerCondition(i cmapi.GenericIssuer, conditionType cmapi.IssuerConditionType, status cmmeta.ConditionStatus, reason, message string) {
	newCondition := cmapi.IssuerCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	nowTime := metav1.NewTime(Clock.Now())
	newCondition.LastTransitionTime = &nowTime

	// Search through existing conditions
	for idx, cond := range i.GetStatus().Conditions {
		// Skip unrelated conditions
		if cond.Type != conditionType {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		} else {
			logf.V(logf.InfoLevel).Infof("Found status change for Issuer %q condition %q: %q -> %q; setting lastTransitionTime to %v", i.GetObjectMeta().Name, conditionType, cond.Status, status, nowTime.Time)
		}

		// Overwrite the existing condition
		i.GetStatus().Conditions[idx] = newCondition
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	i.GetStatus().Conditions = append(i.GetStatus().Conditions, newCondition)
	logf.V(logf.InfoLevel).Infof("Setting lastTransitionTime for Issuer %q condition %q to %v", i.GetObjectMeta().Name, conditionType, nowTime.Time)
}

// CertificateHasCondition will return true if the given Certificate has a
// condition matching the provided CertificateCondition.
// Only the Type and Status field will be used in the comparison, meaning that
// this function will return 'true' even if the Reason, Message and
// LastTransitionTime fields do not match.
func CertificateHasCondition(crt *cmapi.Certificate, c cmapi.CertificateCondition) bool {
	if crt == nil {
		return false
	}
	existingConditions := crt.Status.Conditions
	for _, cond := range existingConditions {
		if c.Type == cond.Type && c.Status == cond.Status {
			return true
		}
	}
	return false
}

func GetCertificateCondition(crt *cmapi.Certificate, conditionType cmapi.CertificateConditionType) *cmapi.CertificateCondition {
	for _, cond := range crt.Status.Conditions {
		if cond.Type == conditionType {
			return &cond
		}
	}
	return nil
}

func GetCertificateRequestCondition(req *cmapi.CertificateRequest, conditionType cmapi.CertificateRequestConditionType) *cmapi.CertificateRequestCondition {
	for _, cond := range req.Status.Conditions {
		if cond.Type == conditionType {
			return &cond
		}
	}
	return nil
}

// SetCertificateCondition will set a 'condition' on the given Certificate.
// - If no condition of the same type already exists, the condition will be
//   inserted with the LastTransitionTime set to the current time.
// - If a condition of the same type and state already exists, the condition
//   will be updated but the LastTransitionTime will not be modified.
// - If a condition of the same type and different state already exists, the
//   condition will be updated and the LastTransitionTime set to the current
//   time.
func SetCertificateCondition(crt *cmapi.Certificate, conditionType cmapi.CertificateConditionType, status cmmeta.ConditionStatus, reason, message string) {
	newCondition := cmapi.CertificateCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	nowTime := metav1.NewTime(Clock.Now())
	newCondition.LastTransitionTime = &nowTime

	// Search through existing conditions
	for idx, cond := range crt.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != conditionType {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		} else {
			logf.V(logf.InfoLevel).Infof("Found status change for Certificate %q condition %q: %q -> %q; setting lastTransitionTime to %v", crt.Name, conditionType, cond.Status, status, nowTime.Time)
		}

		// Overwrite the existing condition
		crt.Status.Conditions[idx] = newCondition
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	crt.Status.Conditions = append(crt.Status.Conditions, newCondition)
	logf.V(logf.InfoLevel).Infof("Setting lastTransitionTime for Certificate %q condition %q to %v", crt.Name, conditionType, nowTime.Time)
}

// RemoteCertificateCondition will remove any condition with this condition type
func RemoveCertificateCondition(crt *cmapi.Certificate, conditionType cmapi.CertificateConditionType) {
	var updatedConditions []cmapi.CertificateCondition

	// Search through existing conditions
	for _, cond := range crt.Status.Conditions {
		// Only add unrelated conditions
		if cond.Type != conditionType {
			updatedConditions = append(updatedConditions, cond)
		}
	}

	crt.Status.Conditions = updatedConditions
}

// SetCertificateRequestCondition will set a 'condition' on the given CertificateRequest.
// - If no condition of the same type already exists, the condition will be
//   inserted with the LastTransitionTime set to the current time.
// - If a condition of the same type and state already exists, the condition
//   will be updated but the LastTransitionTime will not be modified.
// - If a condition of the same type and different state already exists, the
//   condition will be updated and the LastTransitionTime set to the current
//   time.
func SetCertificateRequestCondition(cr *cmapi.CertificateRequest, conditionType cmapi.CertificateRequestConditionType, status cmmeta.ConditionStatus, reason, message string) {
	newCondition := cmapi.CertificateRequestCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	nowTime := metav1.NewTime(Clock.Now())
	newCondition.LastTransitionTime = &nowTime

	// Search through existing conditions
	for idx, cond := range cr.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != conditionType {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		} else {
			logf.V(logf.InfoLevel).Infof("Found status change for CertificateRequest %q condition %q: %q -> %q; setting lastTransitionTime to %v", cr.Name, conditionType, cond.Status, status, nowTime.Time)
		}

		// Overwrite the existing condition
		cr.Status.Conditions[idx] = newCondition
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	cr.Status.Conditions = append(cr.Status.Conditions, newCondition)
	logf.V(logf.InfoLevel).Infof("Setting lastTransitionTime for CertificateRequest %q condition %q to %v", cr.Name, conditionType, nowTime.Time)
}

// CertificateRequestHasCondition will return true if the given
// CertificateRequest has a condition matching the provided
// CertificateRequestCondition.
// Only the Type and Status field will be used in the comparison, meaning that
// this function will return 'true' even if the Reason, Message and
// LastTransitionTime fields do not match.
func CertificateRequestHasCondition(cr *cmapi.CertificateRequest, c cmapi.CertificateRequestCondition) bool {
	if cr == nil {
		return false
	}
	existingConditions := cr.Status.Conditions
	for _, cond := range existingConditions {
		if c.Type == cond.Type && c.Status == cond.Status {
			if c.Reason == "" || c.Reason == cond.Reason {
				return true
			}
		}
	}
	return false
}

// This returns the status reason of a CertificateRequest. The order of reason
// hierarchy is 'Failed' -> 'Ready' -> 'Pending' -> ''
func CertificateRequestReadyReason(cr *cmapi.CertificateRequest) string {
	for _, reason := range []string{
		cmapi.CertificateRequestReasonFailed,
		cmapi.CertificateRequestReasonIssued,
		cmapi.CertificateRequestReasonPending,
	} {
		for _, con := range cr.Status.Conditions {
			if con.Type == cmapi.CertificateRequestConditionReady &&
				con.Reason == reason {
				return reason
			}
		}
	}

	return ""
}

// This returns with the message if the CertificateRequest contains an
// InvalidRequest condition, and returns "" otherwise.
func CertificateRequestInvalidRequestMessage(cr *cmapi.CertificateRequest) string {
	if cr == nil {
		return ""
	}

	for _, con := range cr.Status.Conditions {
		if con.Type == cmapi.CertificateRequestConditionInvalidRequest &&
			con.Status == cmmeta.ConditionTrue {
			return con.Message
		}
	}

	return ""
}

// This returns with true if the CertificateRequest contains an InvalidRequest
// condition, and returns false otherwise.
func CertificateRequestHasInvalidRequest(cr *cmapi.CertificateRequest) bool {
	if cr == nil {
		return false
	}

	for _, con := range cr.Status.Conditions {
		if con.Type == cmapi.CertificateRequestConditionInvalidRequest &&
			con.Status == cmmeta.ConditionTrue {
			return true
		}
	}

	return false
}
