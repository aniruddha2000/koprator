package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"math/rand"

	"github.com/aniruddha2000/koprator/pkg/runtime"
	"github.com/aniruddha2000/koprator/pkg/subscription"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	minWatchTimeout = 5 * time.Minute
	timeoutSeconds  = int64(minWatchTimeout.Seconds() * (rand.Float64() + 1.0))
	masterURL       string
	kubeconfig      string
	addr            = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	start := time.Now()
	klog.Infof("Starting @ %s", start.String())

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*addr, nil))
	}()

	klog.Info("Got watcher client...")

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	klog.Info("Building config from flags...")

	defaultKubernetesClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building watcher clientset: %s", err.Error())
	}

	// Context
	context := context.TODO()

	// Subscription objects
	configMapSubscription := &subscription.ConfigMapSubscribtion{
		Client: defaultKubernetesClientset,
		Ctx:    context,
	}
	podSubscription := &subscription.PodSubscribtion{
		Client:               defaultKubernetesClientset,
		Ctx:                  context,
		ConfigMapSubscribRef: configMapSubscription,
	}

	if err := runtime.RunLoop([]subscription.Subscribtion{
		configMapSubscription,
		podSubscription,
	}); err != nil {
		klog.Fatalf(err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
