package main

import (
	"context"
	"flag"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"time"

	"github.com/aniruddha2000/koprator/pkg/runtime"
	"github.com/aniruddha2000/koprator/pkg/subscription"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	minWatchTimeout = 5 * time.Minute
	addr            = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	masterURL       string
	kubeConfig      string
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
		log.Errorf("Error serving http server at %s: %v", *addr, err)
	}()

	// Logs
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	// Run
	log.Info("Got watcher client...")

	kubernetesCfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		log.Errorf("Error building kubeconfig: %s", err.Error())
	}

	log.Info("Building config from flags...")

	defaultKubernetesClientSet, err := kubernetes.NewForConfig(kubernetesCfg)
	if err != nil {
		log.Errorf("Error building watcher clientset: %s", err.Error())
	}

	// Context
	ctx := context.TODO()

	// Subscription objects
	configMapSubscription := &subscription.ConfigMapSubscription{
		Client: defaultKubernetesClientSet,
	}
	podSubscription := &subscription.PodSubscription{
		Client:                defaultKubernetesClientSet,
		ConfigMapSubscribeRef: configMapSubscription,
	}

	runtime.RunLoop(ctx, []subscription.Subscription{
		configMapSubscription,
		podSubscription,
	})
}

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
