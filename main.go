package main

import (
	"net/http"
	"os"

	"k8s-sd-controller/internal/api"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

func main() {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	labelSelector, err := labels.Parse("my-app.kubernetes.io/exposed=true")
	if err != nil {
		os.Exit(1)
	}

	// Spin up a cluster handle to configure the shared cache engine
	cluster, err := cluster.New(ctrl.GetConfigOrDie(), func(o *cluster.Options) {
		o.Scheme = scheme
		o.Cache = cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.Service{}: {Label: labelSelector},
			},
		}
	})
	if err != nil {
		os.Exit(1)
	}

	// Construct the API layer using the cluster's high-performance cache client
	apiServer := api.NewApiServer(cluster.GetClient())

	ctx := ctrl.SetupSignalHandler()

	// Spin up the cache informer loops in the background
	go func() {
		if err := cluster.Start(ctx); err != nil {
			os.Exit(1)
		}
	}()

	// Blocks and waits for the cache handshake to finish so traffic isn't served empty lists
	if cluster.GetCache().WaitForCacheSync(ctx) {
		_ = http.ListenAndServe(":8080", apiServer.Handler())
	}
}
