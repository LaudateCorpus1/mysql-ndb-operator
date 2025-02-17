// Copyright (c) 2021, Oracle and/or its affiliates.
//
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package ndbtest

import (
	yaml_utils "github.com/mysql/ndb-operator/e2e-tests/utils/yaml"
	"github.com/onsi/gomega"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"strings"
)

const (
	CreateCmd = "create"
	DeleteCmd = "delete"
	ApplyCmd  = "apply"
)

// RunKubectl is a wrapper around framework.RunKubectlInput that additionally expects no error
func RunKubectl(command, namespace, yamlContent string) string {
	if command != CreateCmd && command != DeleteCmd && command != ApplyCmd {
		panic("Unsupported command in getKubectlArgs")
	}

	result, err := framework.RunKubectlInput(namespace, yamlContent, command, "-n", namespace, "-f", "-")
	framework.ExpectNoError(err, "RunKubectl failed with an error")
	klog.V(3).Infof("kubectl executing %s : %s", command, result)
	return result
}

// CreateObjectsFromYaml extracts the resource identified by the
// Kind, Version, Name and creates that in the k8s cluster
// If the ns string is not empty, the resource will be created
// in the namespace pointed by it
func CreateObjectsFromYaml(path, filename string, k8sObjects []yaml_utils.K8sObject, namespace string) {
	// Apply the yaml to k8s
	kubectlInput := yaml_utils.ExtractObjectsFromYaml(path, filename, k8sObjects, namespace)
	gomega.Expect(kubectlInput).NotTo(gomega.BeEmpty())
	result := RunKubectl(ApplyCmd, namespace, kubectlInput)
	klog.V(3).Infof("kubectl executing %s on some objects from %s: %s", "apply", filename, result)
	gomega.Expect(strings.Count(result, "created")).To(gomega.Equal(len(k8sObjects)))
}

// DeleteObjectsFromYaml extracts the resource identified by the
// Kind, Version, Name and deletes them from the k8s cluster
// If the ns string is not empty, the resource will be deleted
// from the namespace pointed by it
func DeleteObjectsFromYaml(path, filename string, k8sObjects []yaml_utils.K8sObject, namespace string) {
	// Apply the yaml to k8s
	kubectlInput := yaml_utils.ExtractObjectsFromYaml(path, filename, k8sObjects, namespace)
	gomega.Expect(kubectlInput).NotTo(gomega.BeEmpty())
	result := RunKubectl(DeleteCmd, namespace, kubectlInput)
	klog.V(3).Infof("kubectl executing %s on some objects from %s: %s", "delete", filename, result)
	gomega.Expect(strings.Count(result, "deleted")).To(gomega.Equal(len(k8sObjects)))
}
