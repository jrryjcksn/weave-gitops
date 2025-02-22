package server

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsRequest) (*pb.ListKustomizationsResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(msg.Namespace)); err != nil {
		return nil, err
	}

	var results []*pb.Kustomization

	for n, l := range clist.Lists() {
		list, ok := l.(*kustomizev1.KustomizationList)
		if !ok {
			continue
		}

		for _, kustomization := range list.Items {
			k, err := types.KustomizationToProto(&kustomization, n)
			if err != nil {
				return nil, fmt.Errorf("converting items: %w", err)
			}

			results = append(results, k)
		}
	}

	return &pb.ListKustomizationsResponse{
		Kustomizations: results,
	}, nil
}

func (cs *coreServer) GetKustomization(ctx context.Context, msg *pb.GetKustomizationRequest) (*pb.GetKustomizationResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	k := &kustomizev1.Kustomization{}
	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := clustersClient.Get(ctx, msg.ClusterName, key, k); err != nil {
		return nil, err
	}

	res, err := types.KustomizationToProto(k, msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("converting kustomization to proto: %w", err)
	}

	return &pb.GetKustomizationResponse{Kustomization: res}, nil
}
