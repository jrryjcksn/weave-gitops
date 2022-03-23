package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/cache"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, cfg CoreServerConfig) error {
	appsServer := NewCoreServer(cfg)
	if err := pb.RegisterCoreHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	return nil
}

const temporarilyEmptyAppName = ""

type coreServer struct {
	pb.UnimplementedCoreServer

	k8s            kube.ClientGetter
	cacheContainer *cache.Container
	logger         logr.Logger
}

type CoreServerConfig struct {
	log         logr.Logger
	RestCfg     *rest.Config
	clusterName string
}

func NewCoreConfig(log logr.Logger, cfg *rest.Config, clusterName string) CoreServerConfig {
	return CoreServerConfig{
		log:         log.WithName("core-server"),
		RestCfg:     cfg,
		clusterName: clusterName,
	}
}

func NewCoreServer(cfg CoreServerConfig) pb.CoreServer {
	cfgGetter := kube.NewImpersonatingConfigGetter(cfg.RestCfg, false)
	k8s := kube.NewDefaultClientGetter(cfgGetter, cfg.clusterName)
	ctx := context.Background()
	cacheContainer := cache.NewContainer(ctx, k8s)

	cacheContainer.Start(ctx)

	return &coreServer{
		k8s:            k8s,
		logger:         cfg.log,
		cacheContainer: cacheContainer,
	}
}

func list(ctx context.Context, k8s client.Client, appName, namespace string, list client.ObjectList, extraOpts ...client.ListOption) error {
	opts := []client.ListOption{
		getMatchingLabels(appName),
		client.InNamespace(namespace),
	}

	opts = append(opts, extraOpts...)
	err := k8s.List(ctx, list, opts...)
	err = wrapK8sAPIError("list resource", err)

	return err
}

func get(ctx context.Context, k8s client.Client, name, namespace string, obj client.Object) error {
	err := k8s.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, obj)
	err = wrapK8sAPIError("get resource", err)

	return err
}

func wrapK8sAPIError(msg string, err error) error {
	if k8serrors.IsUnauthorized(err) {
		return status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
