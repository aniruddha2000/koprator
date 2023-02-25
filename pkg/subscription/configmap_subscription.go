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

type ConfigMapSubscribtion struct {
	watcherInterface    watch.Interface
	Client              kubernetes.Interface
	Ctx                 context.Context
	platformConfig      *PlatformConfig
	paltformConfigPhase watch.EventType
}

var (
	platformConfigMapName                   string = "platform-default-configmap"
	platformConfigMapNamespace              string = "kube-system"
	prometheusPlatfromConfigAnnotationCount        = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "platform_config_annotation_count",
		Help: "This tells us the number of annotations in configmap",
	})
	prometheusPlatformConfigAvailabilityGuage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_config_availability",
		Help: "This tells us weather platform config available",
	}, []string{"configmap_name", "namespace"})
)

type PlatformAnnotations struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type PlatformConfig struct {
	Annotations []PlatformAnnotations `yaml:"annotations"`
}

func isPlatformConfigMap(cm *v1.ConfigMap) (bool, error) {
	if cm == nil {
		return false, errors.New("empty platform configMap")
	}

	if cm.Name == platformConfigMapName {
		return true, nil
	}
	return false, nil
}

func (c *ConfigMapSubscribtion) Reconcile(object runtime.Object, event watch.EventType) {
	cm := object.(*v1.ConfigMap)
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

func (c *ConfigMapSubscribtion) Subscribe() (watch.Interface, error) {
	var err error

	c.watcherInterface, err = c.Client.CoreV1().ConfigMaps("").Watch(c.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("configmap: watcher interface: %v", err)
	}

	return c.watcherInterface, nil
}
