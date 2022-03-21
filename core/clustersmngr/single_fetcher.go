package clustersmngr

import (
	"context"

	"k8s.io/client-go/rest"
)

const DefaultClusterName = "Default"

type singleClusterFetcher struct {
	restConfig *rest.Config
}

func NewSingleClusterFetcher(config *rest.Config) (ClusterFetcher, error) {
	return singleClusterFetcher{
		restConfig: config,
	}, nil
}

func (cf singleClusterFetcher) Fetch(ctx context.Context) ([]Cluster, error) {
	return []Cluster{
		{
			Name:        DefaultClusterName,
			Server:      cf.restConfig.Host,
			BearerToken: cf.restConfig.BearerToken,
			TLSConfig:   cf.restConfig.TLSClientConfig,
		},
	}, nil
}
