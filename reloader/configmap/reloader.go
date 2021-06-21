package configmap

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"sync"
	"time"
)
import "sigs.k8s.io/controller-runtime/pkg/client/config"

type Dir struct {
	namespace string
	name      string
	kube      kubernetes.Interface
	load      func(files map[string][]byte)
	current   map[string][]byte
	mu        sync.Mutex
}

func (d *Dir) LoadOnce() {
	d.load(d.files())
}

func (d *Dir) files() map[string][]byte {
	m := map[string][]byte{}
	d.mu.Lock()
	for name, bytes := range d.current {
		m[name] = bytes
	}
	d.mu.Unlock()
	return m
}

func configMapFiles(cm *corev1.ConfigMap) (files map[string][]byte) {
	if cm == nil {
		return map[string][]byte{}
	}
	m := map[string][]byte{}

	for fileName, content := range cm.Data {
		m[fileName] = []byte(content)
	}
	for fileName, bytes := range cm.BinaryData {
		m[fileName] = bytes
	}
	return m
}

func (d *Dir) OnAdd(obj interface{}) {
	if cm, ok := obj.(*corev1.ConfigMap); ok {
		m := configMapFiles(cm)
		changed := true
		d.mu.Lock()
		if reflect.DeepEqual(m, d.current) {
			changed = false
		} else {
			d.current = m
		}
		d.mu.Unlock()
		if changed {
			d.load(m)
		}
	}
}

func (d *Dir) OnUpdate(oldObj, obj interface{}) {
	d.OnAdd(obj)
}

func (d *Dir) OnDelete(obj interface{}) {
	// ignore
}

func New(namespace string, name string, load func(files map[string][]byte)) (r *Dir, err error) {
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

	cm, err := k.CoreV1().ConfigMaps(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	m := configMapFiles(cm)
	load(m)
	return &Dir{
		namespace: namespace,
		name:      name,
		kube:      k,
		current:   m,
		load:      load,
	}, nil
}

func (d *Dir) Run(ctx context.Context) error {
	factory := informers.NewSharedInformerFactoryWithOptions(d.kube, time.Minute*15,
		informers.WithNamespace(d.namespace), informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector(v1.ObjectNameField, d.name).String()
		}))
	i := factory.Core().V1().ConfigMaps().Informer()
	i.AddEventHandler(d)
	factory.Start(ctx.Done())
	<-ctx.Done()
	return ctx.Err()
}
