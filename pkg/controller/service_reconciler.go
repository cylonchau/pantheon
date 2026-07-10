package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/model"
)

type ServiceReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	PhClient *PantheonClient
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	klog.V(4).Infof("Reconciling Service %s/%s", req.Namespace, req.Name)

	// 1. Fetch Service object
	svc := &corev1.Service{}
	err := r.Client.Get(ctx, req.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			// Service was deleted, clean up matching targets on server
			return reconcile.Result{}, r.deleteServiceTargets(req.Namespace, req.Name)
		}
		return reconcile.Result{}, err
	}

	// 2. Fetch active MonitorRules from Pantheon
	rules, err := r.PhClient.FetchMonitorRules()
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3. Check if Service matches any "service" monitor rule
	matched := false
	var matchingRule model.MonitorRule
	for _, rule := range rules {
		if rule.Type != "service" {
			continue
		}
		if rule.Namespace != "*" && rule.Namespace != "" && rule.Namespace != svc.Namespace {
			continue
		}
		selectorMap := parseKeyValueString(rule.SelectorString)
		selector := labels.SelectorFromSet(selectorMap)
		if selector.Matches(labels.Set(svc.Labels)) {
			matched = true
			matchingRule = rule
			break
		}
	}

	if !matched {
		// Does not match rules, clean up any existing target for this Service
		return reconcile.Result{}, r.deleteServiceTargets(svc.Namespace, svc.Name)
	}

	// 4. Service matches, find matching Pods in K8s
	if len(svc.Spec.Selector) == 0 {
		return reconcile.Result{}, nil
	}

	podList := &corev1.PodList{}
	err = r.Client.List(ctx, podList, client.InNamespace(svc.Namespace), client.MatchingLabels(svc.Spec.Selector))
	if err != nil {
		return reconcile.Result{}, err
	}

	discoveredAddresses := make(map[string]bool)
	discoveredTargets := make([]target.TargetItem, 0)

	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.PodIP == "" {
			continue
		}

		port, err := resolveServicePort(svc, pod, matchingRule.PortName)
		if err != nil {
			klog.V(4).Infof("Skipping pod %s/%s for Service %s: %v", pod.Namespace, pod.Name, svc.Name, err)
			continue
		}

		addr := fmt.Sprintf("%s:%d", pod.Status.PodIP, port)
		discoveredAddresses[addr] = true

		// Build target labels
		targetLabels := map[string]string{
			"kubernetes_namespace":    pod.Namespace,
			"kubernetes_pod_name":     pod.Name,
			"kubernetes_service_name": svc.Name,
			"kubernetes_node_name":    pod.Spec.NodeName,
			"monitor_rule":            matchingRule.Name,
		}
		for k, v := range parseKeyValueString(matchingRule.LabelsString) {
			targetLabels[k] = v
		}
		if matchingRule.DropMetrics != "" {
			targetLabels["drop_metrics"] = matchingRule.DropMetrics
		}

		discoveredTargets = append(discoveredTargets, target.TargetItem{
			Address:    addr,
			MetricPath: matchingRule.MetricPath,
			Labels:     targetLabels,
		})
	}

	// Fetch registered targets on server
	registered, err := r.PhClient.FetchRegisteredTargets()
	if err != nil {
		return reconcile.Result{}, err
	}

	// Clean up registered targets under this Service that are no longer active/discovered
	for _, reg := range registered {
		if reg.Labels["kubernetes_namespace"] == svc.Namespace && reg.Labels["kubernetes_service_name"] == svc.Name {
			cleanAddr := reg.Address
			if idx := strings.Index(cleanAddr, "://"); idx != -1 {
				cleanAddr = cleanAddr[idx+3:]
			}
			if !discoveredAddresses[cleanAddr] {
				klog.Infof("Deleting obsolete Service target ID %d (%s) from server", reg.ID, reg.Address)
				if err := r.PhClient.DeleteTarget(reg.ID); err != nil {
					klog.Errorf("Failed to delete target %d: %v", reg.ID, err)
				}
			}
		}
	}

	// Register or update active targets
	for _, item := range discoveredTargets {
		isNewOrUpdated := true
		for _, reg := range registered {
			cleanAddr := reg.Address
			if idx := strings.Index(cleanAddr, "://"); idx != -1 {
				cleanAddr = cleanAddr[idx+3:]
			}
			if cleanAddr == item.Address && reg.MetricPath == item.MetricPath {
				labelsMatch := true
				for k, v := range item.Labels {
					if reg.Labels[k] != v {
						labelsMatch = false
						break
					}
				}
				if labelsMatch {
					isNewOrUpdated = false
				}
				break
			}
		}

		if isNewOrUpdated {
			klog.Infof("Registering Service target %s for Service %s/%s under rule %q", item.Address, svc.Namespace, svc.Name, matchingRule.Name)
			if err := r.PhClient.RegisterTarget(matchingRule.Name, item); err != nil {
				klog.Errorf("Failed to register target %s: %v", item.Address, err)
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ServiceReconciler) deleteServiceTargets(namespace, name string) error {
	registered, err := r.PhClient.FetchRegisteredTargets()
	if err != nil {
		return err
	}

	for _, reg := range registered {
		if reg.Labels["kubernetes_namespace"] == namespace && reg.Labels["kubernetes_service_name"] == name {
			klog.Infof("Deleting Service target ID %d (%s) from server", reg.ID, reg.Address)
			if err := r.PhClient.DeleteTarget(reg.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveServicePort(svc *corev1.Service, pod *corev1.Pod, portNameOrNum string) (int32, error) {
	// 1. If it's a numeric port, use directly
	var err error
	var num int
	if num, err = strconv.Atoi(portNameOrNum); err == nil {
		return int32(num), nil
	}

	// 2. Look up the port in the Service ports list
	var targetPortStr string
	var targetPortInt int32
	found := false
	for _, p := range svc.Spec.Ports {
		if p.Name == portNameOrNum {
			if p.TargetPort.Type == intstr.Int {
				targetPortInt = p.TargetPort.IntVal
			} else {
				targetPortStr = p.TargetPort.StrVal
			}
			found = true
			break
		}
	}

	if !found {
		// Not found in service spec, attempt direct pod container port resolution
		return resolvePodPort(pod, portNameOrNum)
	}

	// 3. Resolve target port
	if targetPortInt > 0 {
		return targetPortInt, nil
	}
	return resolvePodPort(pod, targetPortStr)
}
