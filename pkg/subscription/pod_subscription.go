package subscription

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type PodSubscribtion struct {
	watcherInterface     watch.Interface
	Client               kubernetes.Interface
	Ctx                  context.Context
	ConfigMapSubscribRef *ConfigMapSubscribtion
}

func (p *PodSubscribtion) applyConfigMapChanges(pod *v1.Pod) {
	if p.ConfigMapSubscribRef != nil {
		if p.ConfigMapSubscribRef.platformConfig != nil {
			updatedPod := pod.DeepCopy()
			klog.Info(updatedPod.Annotations)

			if updatedPod.Annotations == nil {
				updatedPod.Annotations = make(map[string]string, 2)
			}

			for _, annotation := range p.ConfigMapSubscribRef.platformConfig.Annotations {
				updatedPod.Annotations[annotation.Name] = annotation.Value
			}

			_, err := p.Client.CoreV1().Pods(pod.Namespace).Update(p.Ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				klog.Fatal(err.Error())
			}
		}
	}
}

func (p *PodSubscribtion) Reconcile(object runtime.Object, event watch.EventType) {
	pod := object.(*v1.Pod)
	klog.Infof("PodSubscription event type %s for %s", event, pod.Name)

	switch event {
	case watch.Added:
		p.applyConfigMapChanges(pod)
	case watch.Modified:
		p.applyConfigMapChanges(pod)
	}
}

func (p *PodSubscribtion) Subscribe() (watch.Interface, error) {
	var err error

	p.watcherInterface, err = p.Client.CoreV1().Pods("").Watch(p.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return p.watcherInterface, nil
}
