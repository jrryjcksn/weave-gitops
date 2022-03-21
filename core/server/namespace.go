package server

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

const (
	fluxNamespacePartOf   = "flux"
	fluxNamespaceInstance = "flux-system"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

func (as *coreServer) GetFluxNamespace(ctx context.Context, msg *pb.GetFluxNamespaceRequest) (*pb.GetFluxNamespaceResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	nsList := corev1.NamespaceList{}
	options := matchLabel(
		withPartOfLabel(fluxNamespacePartOf),
		withInstanceLabel(fluxNamespaceInstance),
	)

	if err = k8s.List(ctx, &nsList, &options); err != nil {
		return nil, doClientError(err)
	}

	if len(nsList.Items) == 0 {
		return nil, ErrNamespaceNotFound
	}

	return &pb.GetFluxNamespaceResponse{Name: nsList.Items[0].Name}, nil
}

func (as *coreServer) ListNamespaces(ctx context.Context, msg *pb.ListNamespacesRequest) (*pb.ListNamespacesResponse, error) {
	pool := clustersmngr.ClientsPoolFromCtx(ctx)

	if pool == nil {
		return nil, errors.New("getting clients pool from context: pool was nil")
	}

	// Only doing on-cluster things for now.
	namespaces, err := as.nsLister.ListNamespaces(ctx, clustersmngr.DefaultClusterName)
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}

	clients := pool.Clients()
	if clients == nil {
		return nil, errors.New("no pool clients")
	}

	client, ok := clients[clustersmngr.DefaultClusterName]
	if !ok {
		return nil, fmt.Errorf("no client found in pool for %q", clustersmngr.DefaultClusterName)
	}

	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	for _, ns := range namespaces {
		ok, err := userCanUseNamespace(ctx, client.RestConfig(), ns)
		if err != nil {
			return nil, fmt.Errorf("user namespace access: %w", err)
		}

		if ok {
			response.Namespaces = append(response.Namespaces, types.NamespaceToProto(ns))
		}
	}

	return response, nil
}

func userCanUseNamespace(ctx context.Context, cfg *rest.Config, ns corev1.Namespace) (bool, error) {
	necessaryRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"secrets", "pods", "events", "namespaces"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get", "list"},
		},
	}

	auth, err := newAuthClient(cfg)
	if err != nil {
		return false, err
	}

	sar := &authorizationv1.SelfSubjectRulesReview{
		Spec: authorizationv1.SelfSubjectRulesReviewSpec{
			Namespace: ns.Name,
		},
	}

	authRes, err := auth.SelfSubjectRulesReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}

	return hasAllRules(authRes.Status, necessaryRules), nil
}

func hasAllRules(status authorizationv1.SubjectRulesReviewStatus, rules []rbacv1.PolicyRule) bool {
	for _, rule := range rules {
		if !hasRule(rule, status.ResourceRules) {
			return false
		}
	}

	return true
}

func hasRule(rule rbacv1.PolicyRule, collection []authorizationv1.ResourceRule) bool {
	for _, rule := range collection {
		for _, apiGroup := range rule.APIGroups {
			if sort.SearchStrings(rule.APIGroups, apiGroup) == -1 {
				return false
			}
		}

		for _, resource := range rule.Resources {
			if sort.SearchStrings(rule.Resources, resource) == -1 {
				return false
			}
		}

		for _, verb := range rule.Verbs {
			if sort.SearchStrings(rule.Verbs, verb) == -1 {
				return false
			}
		}
	}

	return true
}

func newAuthClient(cfg *rest.Config) (typedauth.AuthorizationV1Interface, error) {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("making clientset: %w", err)
	}

	return cs.AuthorizationV1(), nil
}
