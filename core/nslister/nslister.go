package nslister

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

type NSLister interface {
	ListNamespaces(ctx context.Context, clusterName string) ([]corev1.Namespace, error)
}

type simpleLister struct {
}

func NewNSLister() NSLister {
	return simpleLister{}
}

func (sl simpleLister) ListNamespaces(ctx context.Context, clusterName string) ([]corev1.Namespace, error) {
	// Using a non-impersonated, on-cluster client here
	_, k8s, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("creating admin client: %w", err)
	}

	nsList := corev1.NamespaceList{}

	if err := k8s.List(ctx, &nsList); err != nil {
		return nil, fmt.Errorf("listing namespaces for cluster %q: %w", clusterName, err)
	}

	return nsList.Items, nil
}
