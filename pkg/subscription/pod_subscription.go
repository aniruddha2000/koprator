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
	watcherInterface watch.Interface
	Client           kubernetes.Interface
	Ctx              context.Context
}

func (p *PodSubscribtion) Reconcile(object runtime.Object, event watch.EventType) {
	pod := object.(*v1.Pod)
	klog.Infof("PodSubscription event type %s for %s", event, pod.Name)

	switch event {
	case watch.Added:
		if _, ok := pod.Annotations["type"]; !ok {
			updatedPod := pod.DeepCopy()
			updatedPod.Annotations["type"] = "operator"

			_, err := p.Client.CoreV1().Pods(pod.Namespace).Update(p.Ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				klog.Fatal(err.Error())
			}
		}
	case watch.Modified:
		if pod.Annotations["type"] == "operator" {
			klog.Info("This could be some custome behaviour beyond the CRUD.")
		}
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
