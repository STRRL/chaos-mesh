// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package watcher

import (
	"context"
	"html/template"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/pkg/metrics"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
)

var restClusterConfig = ctrlconfig.GetConfig
var kubernetesNewForConfig = kubernetes.NewForConfig

const (
	serviceAccountNamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	templateItemKey                 = "data"
)

// ErrWatchChannelClosed should restart watcher
var ErrWatchChannelClosed = errors.New("watcher channel has closed")

// K8sConfigMapWatcher is a struct that connects to the API and collects, parses, and emits sidecar configurations
type K8sConfigMapWatcher struct {
	Config
	client  k8sv1.CoreV1Interface
	metrics *metrics.ChaosControllerManagerMetricsCollector
	logger  logr.Logger
}

// TODO: replace New with a Constructor and a wrapper

// New creates a new K8sConfigMapWatcher
func New(cfg Config, metrics *metrics.ChaosControllerManagerMetricsCollector, logger logr.Logger) (*K8sConfigMapWatcher, error) {
	c := K8sConfigMapWatcher{Config: cfg, metrics: metrics, logger: logger}
	if strings.TrimSpace(c.TemplateNamespace) == "" {
		// ENHANCEMENT: support downward API/env vars instead? https://github.com/kubernetes/kubernetes/blob/release-1.0/docs/user-guide/downward-api.md
		// load from file on disk for serviceaccount: /var/run/secrets/kubernetes.io/serviceaccount/namespace
		nsBytes, err := os.ReadFile(serviceAccountNamespaceFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, errors.Wrapf(err, "service account namespace file %s does not exist", serviceAccountNamespaceFilePath)
			}
			return nil, err
		}
		ns := strings.TrimSpace(string(nsBytes))
		if ns != "" {
			c.TemplateNamespace = ns
			logger.Info("Inferred ConfigMap",
				"template namespace", c.TemplateNamespace, "filepath", serviceAccountNamespaceFilePath)
		} else {
			return nil, errors.New("can not found namespace. maybe you should specify --template-namespace if you are running outside of kubernetes")
		}
	}

	logger.Info("Creating Kubernetes client to talk to the api-server")
	k8sConfig, err := restClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetesNewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	c.client = clientset.CoreV1()
	if err = validate(&c); err != nil {
		return nil, errors.Wrap(err, "validate K8sConfigMapWatcher")
	}
	logger.Info("Created ConfigMap watcher",
		"apiserver", k8sConfig.Host, "template namespaces", c.TemplateNamespace,
		"template labels", c.TemplateLabels, "config labels", c.ConfigLabels)
	return &c, nil
}

func validate(c *K8sConfigMapWatcher) error {
	if c == nil {
		return errors.New("configmap watcher was nil")
	}
	if c.TemplateNamespace == "" {
		return errors.New("namespace is empty")
	}
	if c.TemplateLabels == nil {
		return errors.New("template labels was an uninitialized map")
	}
	if c.ConfigLabels == nil {
		return errors.New("config labels was an uninitialized map")
	}
	if c.client == nil {
		return errors.New("k8s client was not setup properly")
	}
	return nil
}

// Watch watches for events impacting watched ConfigMaps and emits their events across a channel
func (c *K8sConfigMapWatcher) Watch(notifyMe chan<- interface{}, stopCh <-chan struct{}) error {
	c.logger.Info("Watching for ConfigMaps for changes",
		"template namespace", c.TemplateNamespace, "labels", c.ConfigLabels)
	templateWatcher, err := c.client.ConfigMaps(c.TemplateNamespace).Watch(
		// FIXME: get context from parameter
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: mapStringStringToLabelSelector(c.TemplateLabels),
		})
	if err != nil {
		return errors.Wrap(err, "create template configmap watcher")
	}

	targetNamespace := ""
	if !c.Config.ClusterScoped {
		targetNamespace = c.TargetNamespace
	}

	configWatcher, err := c.client.ConfigMaps(targetNamespace).Watch(
		// FIXME: get context from parameter
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: mapStringStringToLabelSelector(c.ConfigLabels),
		})
	if err != nil {
		return errors.Wrap(err, "create configmap watcher")
	}
	defer func() {
		configWatcher.Stop()
		templateWatcher.Stop()
	}()
	for {
		select {
		case e, ok := <-templateWatcher.ResultChan():
			// channel may closed caused by HTTP timeout, should restart watcher
			// detail at https://github.com/kubernetes/client-go/issues/334
			if !ok {
				c.logger.V(5).Info("channel has closed, will restart watcher")
				return ErrWatchChannelClosed
			}
			if e.Type == watch.Error {
				return apierrs.FromObject(e.Object)
			}
			c.logger.V(3).Info("type", e.Type, "kind", e.Object.GetObjectKind())
			switch e.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				fallthrough
			case watch.Deleted:
				// signal reconciliation of all InjectionConfigs
				c.logger.V(3).Info("Signalling event received from watch channel",
					"type", e.Type, "kind", e.Object.GetObjectKind())
				notifyMe <- struct{}{}
			default:
				c.logger.Error(nil, "got unsupported event! skipping", "type", e.Type, "kind", e.Object.GetObjectKind())
			}
		case e, ok := <-configWatcher.ResultChan():
			// channel may closed caused by HTTP timeout, should restart watcher
			// detail at https://github.com/kubernetes/client-go/issues/334
			if !ok {
				c.logger.V(5).Info("channel has closed, will restart watcher")
				return ErrWatchChannelClosed
			}
			if e.Type == watch.Error {
				return apierrs.FromObject(e.Object)
			}
			c.logger.V(3).Info("type", e.Type, "kind", e.Object.GetObjectKind())
			switch e.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				fallthrough
			case watch.Deleted:
				// signal reconciliation of all InjectionConfigs
				c.logger.V(3).Info("Signalling event received from watch channel",
					"type", e.Type, "kind", e.Object.GetObjectKind())
				notifyMe <- struct{}{}
			default:
				c.logger.Error(nil, "got unsupported event! skipping", "type", e.Type, "kind", e.Object.GetObjectKind())
			}
			// events! yay!
		case <-stopCh:
			c.logger.V(2).Info("Stopping configmap watcher, context indicated we are done")
			// clean up, we cancelled the context, so stop the watch
			return nil
		}
	}
}

func mapStringStringToLabelSelector(m map[string]string) string {
	// https://github.com/kubernetes/apimachinery/issues/47
	return labels.Set(m).String()
}

// GetInjectionConfigs fetches all matching ConfigMaps
func (c *K8sConfigMapWatcher) GetInjectionConfigs() (map[string][]*config.InjectionConfig, error) {
	templates, err := c.GetTemplates()
	if err != nil {
		return nil, err
	}

	configs, err := c.GetConfigs()
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 || len(configs) == 0 {
		c.logger.Info("cannot get injection configs")
		return nil, nil
	}

	injectionConfigs := make(map[string][]*config.InjectionConfig)
	if c.metrics != nil {
		c.metrics.InjectionConfigs.Reset()
	}
	for _, conf := range configs {
		temp, ok := templates[conf.Template]
		if !ok {
			c.logger.Error(errors.New("cannot find the specified template"), "",
				"template", conf.Template, "namespace", conf.Namespace, "config", conf.Name)
			if c.metrics != nil {
				c.metrics.TemplateNotExist.WithLabelValues(conf.Namespace, conf.Template).Inc()
			}
			continue
		}
		yamlTemp, err := template.New("").Parse(temp)
		if err != nil {
			c.logger.Error(err, "failed to parse template",
				"template", conf.Template, "config", conf.Name)
			continue
		}

		result, err := renderTemplateWithArgs(yamlTemp, conf.Arguments)
		if err != nil {
			c.logger.Error(err, "failed to render template",
				"template", conf.Template, "config", conf.Name)
			continue
		}

		var injectConfig config.InjectionConfig
		if err := yaml.Unmarshal(result, &injectConfig); err != nil {
			c.logger.Error(err, "failed to unmarshal injection config", "injection config", string(result))
			continue
		}

		injectConfig.Selector = conf.Selector
		injectConfig.Name = conf.Name
		if _, ok := injectionConfigs[conf.Namespace]; !ok {
			injectionConfigs[conf.Namespace] = make([]*config.InjectionConfig, 0)
		}
		injectionConfigs[conf.Namespace] = append(injectionConfigs[conf.Namespace], &injectConfig)
		if c.metrics != nil {
			c.metrics.InjectionConfigs.WithLabelValues(conf.Namespace, conf.Template).Inc()
		}
	}

	return injectionConfigs, nil
}

// GetTemplates returns a map of common templates
func (c *K8sConfigMapWatcher) GetTemplates() (map[string]string, error) {
	c.logger.Info("Fetching Template Configs...")
	templateList, err := c.client.ConfigMaps(c.TemplateNamespace).List(
		// FIXME: get context from parameter
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: mapStringStringToLabelSelector(c.TemplateLabels),
		})
	if err != nil {
		return nil, err
	}

	c.logger.Info("Fetched templates", "templates count", len(templateList.Items))
	templates := make(map[string]string, len(templateList.Items))
	for _, temp := range templateList.Items {
		templates[temp.Name] = temp.Data[templateItemKey]
	}
	if c.metrics != nil {
		c.metrics.SidecarTemplates.Set(float64(len(templates)))
	}
	return templates, nil
}

// GetConfigs returns the list of template args config
func (c *K8sConfigMapWatcher) GetConfigs() ([]*config.TemplateArgs, error) {
	c.logger.Info("Fetching Configs...")
	// List all the configs with the required label selector
	configList, err := c.client.ConfigMaps("").List(
		// FIXME: get context from parmeter
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: mapStringStringToLabelSelector(c.ConfigLabels),
		})
	if err != nil {
		return nil, err
	}

	c.logger.Info("Fetched configs", "configs count", len(configList.Items))
	if c.metrics != nil {
		c.metrics.ConfigTemplates.Reset()
	}
	configSet := make(map[string]map[string]struct{})
	result := make([]*config.TemplateArgs, 0)
	for _, item := range configList.Items {
		for _, payload := range item.Data {
			conf, err := config.LoadTemplateArgs(strings.NewReader(payload))
			if err != nil {
				c.logger.Error(err, "failed to load template args", "payload", payload)
				if c.metrics != nil {
					c.metrics.TemplateLoadError.Inc()
				}
				continue
			}
			conf.Namespace = item.Namespace
			if _, ok := configSet[conf.Namespace]; !ok {
				configSet[conf.Namespace] = make(map[string]struct{})
			}
			if _, ok := configSet[conf.Namespace][conf.Name]; ok {
				c.logger.Error(errors.New("duplicate config name"), "",
					"namespace", conf.Namespace, "name", conf.Name)
				if c.metrics != nil {
					c.metrics.ConfigNameDuplicate.WithLabelValues(conf.Namespace, conf.Name).Inc()
				}
				continue
			}
			configSet[conf.Namespace][conf.Name] = struct{}{}
			if c.metrics != nil {
				c.metrics.ConfigTemplates.WithLabelValues(conf.Namespace, conf.Template).Inc()
			}
			result = append(result, conf)
		}
	}
	return result, nil
}
