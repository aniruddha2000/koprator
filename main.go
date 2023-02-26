package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/aniruddha2000/koprator/pkg/runtime"
	"github.com/aniruddha2000/koprator/pkg/subscription"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	minWatchTimeout = 5 * time.Minute
	masterURL       string
	kubeconfig      string
	addr            = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
)

func main() {
	flag.Parse()

	start := time.Now()
	log.Infof("Starting @ %s", start.String())

	// Metrics
	go func() {
		server := http.Server{
			Addr:              *addr,
			Handler:           promhttp.Handler(),
			ReadHeaderTimeout: minWatchTimeout,
		}
		err := server.ListenAndServe()
		log.Fatalf("Error serving http server at %s: %v", *addr, err)
	}()

	// Logs
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	// Run
	log.Info("Got watcher client...")

	kubernetesCfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	log.Info("Building config from flags...")

	defaultKubernetesClientset, err := kubernetes.NewForConfig(kubernetesCfg)
	if err != nil {
		log.Fatalf("Error building watcher clientset: %s", err.Error())
	}

	// Context
	ctx := context.TODO()

	// Subscription objects
	configMapSubscription := &subscription.ConfigMapSubscribtion{
		Client: defaultKubernetesClientset,
	}
	podSubscription := &subscription.PodSubscribtion{
		Client:               defaultKubernetesClientset,
		ConfigMapSubscribRef: configMapSubscription,
	}

	if err := runtime.RunLoop(ctx, []subscription.Subscription{
		configMapSubscription,
		podSubscription,
	}); err != nil {
		log.Fatalf("Runloop error: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
