// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
//
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/mysql/ndb-operator/pkg/generated/clientset/versioned/typed/ndbcontroller/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeMysqlV1alpha1 struct {
	*testing.Fake
}

func (c *FakeMysqlV1alpha1) Ndbs(namespace string) v1alpha1.NdbInterface {
	return &FakeNdbs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeMysqlV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
