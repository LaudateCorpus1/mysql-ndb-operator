// Copyright (c) 2021, Oracle and/or its affiliates.
//
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package controllers

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	k8s "k8s.io/client-go/kubernetes"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/admissionregistration"
)

// WebhookConfigController defines a webhookConfig control interface
type WebhookConfigController interface {
	UpdateWebhookConfigCertificate(ctx context.Context, labelSelector string, cert []byte) bool
}

// validatingWebhookConfigController implements WebhookConfigController for ValidatingWebhookConfigs
type validatingWebhookConfigController struct {
	vwcInterface admissionregistrationv1.ValidatingWebhookConfigurationInterface
}

// NewValidatingWebhookConfigController creates and returns a new validatingWebhookConfigController
func NewValidatingWebhookConfigController(client k8s.Interface) WebhookConfigController {
	return &validatingWebhookConfigController{
		vwcInterface: client.AdmissionregistrationV1().ValidatingWebhookConfigurations(),
	}
}

// UpdateWebhookConfigCertificate updates the webhooks with the given TLS certificate data
func (c *validatingWebhookConfigController) UpdateWebhookConfigCertificate(
	ctx context.Context, labelSelector string, cert []byte) bool {
	// Get all validating configs with matching label
	vwcList, err := c.vwcInterface.List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		klog.Errorf("Failed to retrieve validating webhook config list : %s", err)
		return false
	}

	// Update them one by one
	for _, vwc := range vwcList.Items {
		// Make a copy of the webhook config to update
		newVwc := vwc.DeepCopy()

		// Update all webhooks' CA bundle
		for i := range newVwc.Webhooks {
			newVwc.Webhooks[i].ClientConfig.CABundle = cert
		}

		// Prepare a patch to be applied
		existingJSON, err := json.Marshal(vwc)
		if err != nil {
			klog.Error("Failed to encode existing validating webhook config : ", err)
			return false
		}
		updatedJSON, err := json.Marshal(newVwc)
		if err != nil {
			klog.Error("Failed to encode updated validating webhook config : ", err)
			return false
		}
		patch, err := strategicpatch.CreateTwoWayMergePatch(
			existingJSON, updatedJSON, admissionregistration.ValidatingWebhookConfiguration{})
		if err != nil {
			klog.Error("Failed to generate the patch to be applied : ", err)
			return false
		}

		// Apply the patch
		if newVwc, err = c.vwcInterface.Patch(
			ctx, newVwc.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
			klog.Errorf("Failed to patch validating webhook config %s : %s",
				newVwc.Name, err)
			return false
		}

		klog.Infof("Successfully updated the validatingWebhookConfig '%s' with the certificate", newVwc.Name)
	}

	return true
}
