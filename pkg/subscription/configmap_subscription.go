package subscription

import (
	"context"
	"errors"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type ConfigMapSubscribtion struct {
	watcherInterface             watch.Interface
	Client                       kubernetes.Interface
	Ctx                          context.Context
	platformConfigMapAnnotations *PlatformConfig
	paltformConfigPhase          watch.EventType
}

const (
	platforConfigMapName      string = "platform-default-configmap"
	platforConfigMapNamespace string = "kube-system"
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

	if cm.Name == platforConfigMapName {
		return true, nil
	}
	return false, nil
}

func (c *ConfigMapSubscribtion) Reconcile(object runtime.Object, event watch.EventType) {
	cm := object.(*v1.ConfigMap)
	c.paltformConfigPhase = event

	if ok, err := isPlatformConfigMap(cm); !ok {
		if err != nil {
			klog.Error(err)
		}
		return
	}
	klog.Infof("ConfigMapSubscription event type %s for %s", event, cm.Name)

	switch event {
	case watch.Added:
		rawDefaultString := cm.Data["platform-default"]

		var unmarshalledData PlatformConfig
		err := yaml.Unmarshal([]byte(rawDefaultString), &unmarshalledData)
		if err != nil {
			klog.Error(err)
			return
		}
		c.platformConfigMapAnnotations = &unmarshalledData

	case watch.Deleted:
		c.platformConfigMapAnnotations = nil

	case watch.Modified:
	}
}

func (c *ConfigMapSubscribtion) Subscribe() (watch.Interface, error) {
	var err error

	c.watcherInterface, err = c.Client.CoreV1().ConfigMaps("").Watch(c.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return c.watcherInterface, nil
}
