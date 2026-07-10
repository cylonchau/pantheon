package controller

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/model"
)

type PodReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	PhClient *PantheonClient
}

func (r *PodReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	klog.V(4).Infof("Reconciling Pod %s/%s", req.Namespace, req.Name)

	// 1. Fetch Pod object
	pod := &corev1.Pod{}
	err := r.Client.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			// Pod was deleted, clean up matching targets on server
			return reconcile.Result{}, r.deletePodTargets(req.Namespace, req.Name)
		}
		return reconcile.Result{}, err
	}

	// 2. Fetch active MonitorRules from Pantheon
	rules, err := r.PhClient.FetchMonitorRules()
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3. Check if Pod matches any "pod" monitor rule
	matched := false
	var matchingRule model.MonitorRule
	for _, rule := range rules {
		if rule.Type != "pod" {
			continue
		}
		if rule.Namespace != "*" && rule.Namespace != "" && rule.Namespace != pod.Namespace {
			continue
		}
		selectorMap := parseKeyValueString(rule.SelectorString)
		selector := labels.SelectorFromSet(selectorMap)
		if selector.Matches(labels.Set(pod.Labels)) {
			matched = true
			matchingRule = rule
			break
		}
	}

	if !matched {
		// Does not match rules, clean up any existing target for this Pod
		return reconcile.Result{}, r.deletePodTargets(pod.Namespace, pod.Name)
	}

	// 4. Pod matches rule, check IP and port
	if pod.Status.PodIP == "" {
		klog.V(4).Infof("Pod %s/%s matched rule, but has no IP yet", pod.Namespace, pod.Name)
		return reconcile.Result{}, nil
	}

	port, err := resolvePodPort(pod, matchingRule.PortName)
	if err != nil {
		klog.Warningf("Failed to resolve port %s for pod %s/%s: %v", matchingRule.PortName, pod.Namespace, pod.Name, err)
		return reconcile.Result{}, nil // Skip until config or pod changes
	}

	addr := fmt.Sprintf("%s:%d", pod.Status.PodIP, port)

	// Build target labels
	targetLabels := map[string]string{
		"kubernetes_namespace": pod.Namespace,
		"kubernetes_pod_name":  pod.Name,
		"kubernetes_node_name": pod.Spec.NodeName,
		"monitor_rule":         matchingRule.Name,
	}
	for k, v := range parseKeyValueString(matchingRule.LabelsString) {
		targetLabels[k] = v
	}
	if matchingRule.DropMetrics != "" {
		targetLabels["drop_metrics"] = matchingRule.DropMetrics
	}

	item := target.TargetItem{
		Address:    addr,
		MetricPath: matchingRule.MetricPath,
		Labels:     targetLabels,
	}

	// Compare with registered targets
	registered, err := r.PhClient.FetchRegisteredTargets()
	if err != nil {
		return reconcile.Result{}, err
	}

	isNewOrUpdated := true
	for _, reg := range registered {
		cleanAddr := reg.Address
		if idx := strings.Index(cleanAddr, "://"); idx != -1 {
			cleanAddr = cleanAddr[idx+3:]
		}
		if cleanAddr == addr && reg.MetricPath == item.MetricPath {
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
		klog.Infof("Registering Pod target %s (%s/%s) under rule %q", addr, pod.Namespace, pod.Name, matchingRule.Name)
		if err := r.PhClient.RegisterTarget(matchingRule.Name, item); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *PodReconciler) deletePodTargets(namespace, name string) error {
	registered, err := r.PhClient.FetchRegisteredTargets()
	if err != nil {
		return err
	}

	for _, reg := range registered {
		if reg.Labels["kubernetes_namespace"] == namespace && reg.Labels["kubernetes_pod_name"] == name {
			klog.Infof("Deleting Pod target ID %d (%s) from server", reg.ID, reg.Address)
			if err := r.PhClient.DeleteTarget(reg.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
