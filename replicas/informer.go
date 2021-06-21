package replicas

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"sync/atomic"
	"time"
)
import "sigs.k8s.io/controller-runtime/pkg/client/config"

type Replicas struct {
	namespace string
	name      string
	kube      kubernetes.Interface
	num       int32
	onUpdate  func(num int32)
}

func (r *Replicas) OnAdd(obj interface{}) {
	if ep, ok := obj.(*corev1.Endpoints); ok {
		num := 0
		for _, subset := range ep.Subsets {
			num = len(subset.Addresses)
			break
		}
		if atomic.LoadInt32(&r.num) != int32(num) {
			atomic.StoreInt32(&r.num, int32(num))
			r.onUpdate(int32(num))
		}
	}
}

func (r *Replicas) OnUpdate(oldObj, obj interface{}) {
	r.OnAdd(obj)
}

func (r *Replicas) OnDelete(obj interface{}) {
	// ignore
}

func New(namespace string, name string, onUpdate func(replicas int32)) (r *Replicas, err error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	k, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	endpoints, err := k.CoreV1().Endpoints(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	num := 0
	for _, subset := range endpoints.Subsets {
		num = len(subset.Addresses)
		break
	}
	onUpdate(int32(num))
	return &Replicas{
		namespace: namespace,
		name:      name,
		kube:      k,
		num:       int32(num),
		onUpdate:  onUpdate,
	}, nil
}

func (r *Replicas) Get() int {
	return int(atomic.LoadInt32(&r.num))
}

func (r *Replicas) Run(ctx context.Context) error {
	factory := informers.NewSharedInformerFactoryWithOptions(r.kube, time.Minute*15,
		informers.WithNamespace(r.namespace), informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector(v1.ObjectNameField, r.name).String()
		}))
	i := factory.Core().V1().Endpoints().Informer()
	i.AddEventHandler(r)
	factory.Start(ctx.Done())
	<-ctx.Done()
	return ctx.Err()
}
