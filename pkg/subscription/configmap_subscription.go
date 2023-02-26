package subscription

import (
	"context"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// ConfigMapSubscribtion defines the attributes for the ConfigMap object reconciliation.
type ConfigMapSubscribtion struct {
	watcherInterface    watch.Interface
	Client              kubernetes.Interface
	platformConfig      *PlatformConfig
	paltformConfigPhase watch.EventType
}

var (
	platformConfigMapName                   = "platform-default-configmap"
	platformConfigMapNamespace              = "kube-system"
	prometheusPlatfromConfigAnnotationCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "platform_config_annotation",
		Help: "This tells us the number of annotations in configmap",
	})
	prometheusPlatformConfigAvailabilityGuage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_config_availability",
		Help: "This tells us weather platform config available",
	}, []string{"configmap_name", "namespace"})
)

// PlatformAnnotations define the annotation structure for the ConfigMap.
type PlatformAnnotations struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// PlatformConfig define the list of annotations in the ConfigMap.
type PlatformConfig struct {
	Annotations []PlatformAnnotations `yaml:"annotations"`
}

func isPlatformConfigMap(cm *v1.ConfigMap) (bool, error) {
	if cm == nil {
		return false, errors.New("empty platform configMap")
	}

	if cm.Name == platformConfigMapName && cm.Namespace == platformConfigMapNamespace {
		return true, nil
	}
	return false, nil
}

// Reconcile gets the ConfigMap annotations in a Add event and store it in the platformConfig attribute
// and make it nil in the Delete event.
func (c *ConfigMapSubscribtion) Reconcile(ctx context.Context, object runtime.Object, event watch.EventType) {
	cm, ok := object.(*v1.ConfigMap)
	if !ok {
		log.Errorf("Want %v but got %v", v1.ConfigMap{}.Kind, cm.Kind)
	}

	c.paltformConfigPhase = event

	if ok, err := isPlatformConfigMap(cm); !ok {
		if err != nil {
			log.Errorf("Reconcile configmap: isPlatformConfigMap: %v", err)
		}
		return
	}
	log.WithFields(log.Fields{
		"namespace": cm.Namespace,
	}).Infof("ConfigMapSubscription event type %s for %s", event, cm.Name)

	switch event {
	case watch.Added:
		rawDefaultString := cm.Data["platform-default"]

		var unmarshalledData PlatformConfig
		err := yaml.Unmarshal([]byte(rawDefaultString), &unmarshalledData)
		if err != nil {
			log.Errorf("Reconcile configmap: yaml unmarshal: %v", err)
			return
		}
		c.platformConfig = &unmarshalledData
		prometheusPlatformConfigAvailabilityGuage.WithLabelValues(cm.Name, cm.Namespace).Set(float64(1))
		prometheusPlatfromConfigAnnotationCount.Set(float64(len(c.platformConfig.Annotations)))

	case watch.Deleted:
		c.platformConfig = nil
		prometheusPlatformConfigAvailabilityGuage.WithLabelValues(cm.Name, cm.Namespace).Set(float64(0))
		prometheusPlatfromConfigAnnotationCount.Set(0)

	case watch.Modified:
	}
}

// Subscribe returns watcher Interface of the ConfigMap object on all namespace.
func (c *ConfigMapSubscribtion) Subscribe(ctx context.Context) (watch.Interface, error) {
	var err error

	c.watcherInterface, err = c.Client.CoreV1().ConfigMaps("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("configmap: watcher interface: %w", err)
	}

	return c.watcherInterface, nil
}
