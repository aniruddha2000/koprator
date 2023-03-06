package subscription

import (
	"context"
	"errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// ConfigMapSubscription defines the attributes for the ConfigMap object reconciliation.
type ConfigMapSubscription struct {
	Client              kubernetes.Interface
	platformConfig      *PlatformConfig
	platformConfigPhase cache.DeltaType
}

var (
	platformConfigMapName                   = "platform-default-configmap"
	platformConfigMapNamespace              = "kube-system"
	prometheusPlatformConfigAnnotationCount = promauto.NewGauge(prometheus.GaugeOpts{
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
func (c *ConfigMapSubscription) Reconcile(ctx context.Context, object interface{}, event cache.DeltaType) {
	cm, ok := object.(*v1.ConfigMap)
	if !ok {
		log.Errorf("Want %v but got %v", v1.ConfigMap{}.Kind, cm.Kind)
	}

	c.platformConfigPhase = event

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
	case cache.Added:
		rawDefaultString := cm.Data["platform-default"]

		var unmarshalledData PlatformConfig
		err := yaml.Unmarshal([]byte(rawDefaultString), &unmarshalledData)
		if err != nil {
			log.Errorf("Reconcile configmap: yaml unmarshal: %v", err)
			return
		}
		c.platformConfig = &unmarshalledData
		prometheusPlatformConfigAvailabilityGuage.WithLabelValues(cm.Name, cm.Namespace).Set(float64(1))
		prometheusPlatformConfigAnnotationCount.Set(float64(len(c.platformConfig.Annotations)))

	case cache.Deleted:
		c.platformConfig = nil
		prometheusPlatformConfigAvailabilityGuage.WithLabelValues(cm.Name, cm.Namespace).Set(float64(0))
		prometheusPlatformConfigAnnotationCount.Set(0)
	}
}

// Subscribe returns Informer factory of the ConfigMap object on all namespace.
func (c *ConfigMapSubscription) Subscribe() (informers.SharedInformerFactory, cache.SharedIndexInformer) {
	informer := informers.NewSharedInformerFactory(c.Client, 10*time.Second)
	cmInformer := informer.Core().V1().ConfigMaps().Informer()

	informer.Start(wait.NeverStop)
	informer.WaitForCacheSync(wait.NeverStop)

	return informer, cmInformer
}
