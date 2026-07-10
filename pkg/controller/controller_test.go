package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apitarget "github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/model"
)

// parseKeyValueString

func TestParseKeyValueString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "Single key-value pair",
			input: "app=portal",
			expected: map[string]string{
				"app": "portal",
			},
		},
		{
			name:  "Multiple key-value pairs with whitespace",
			input: "app=portal , env=prod, team = sre",
			expected: map[string]string{
				"app":  "portal",
				"env":  "prod",
				"team": "sre",
			},
		},
		{
			name:     "Invalid pairs ignored",
			input:    "app=portal,invalidpair,env=prod",
			expected: map[string]string{
				"app": "portal",
				"env": "prod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := parseKeyValueString(tt.input)
			if !reflect.DeepEqual(output, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, output)
			}
		})
	}
}

// resolvePodPort

func TestResolvePodPort(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Ports: []corev1.ContainerPort{
						{Name: "http", ContainerPort: 8080},
						{Name: "grpc", ContainerPort: 9090},
					},
				},
			},
		},
	}

	t.Run("Numeric port string", func(t *testing.T) {
		port, err := resolvePodPort(pod, "9000")
		assert.NoError(t, err)
		assert.Equal(t, int32(9000), port)
	})

	t.Run("Named port found", func(t *testing.T) {
		port, err := resolvePodPort(pod, "http")
		assert.NoError(t, err)
		assert.Equal(t, int32(8080), port)
	})

	t.Run("Named port grpc found", func(t *testing.T) {
		port, err := resolvePodPort(pod, "grpc")
		assert.NoError(t, err)
		assert.Equal(t, int32(9090), port)
	})

	t.Run("Named port not found", func(t *testing.T) {
		_, err := resolvePodPort(pod, "metrics")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in pod containers")
	})
}

// resolveServicePort

func TestResolveServicePort(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Ports: []corev1.ContainerPort{
						{Name: "metrics", ContainerPort: 9100},
					},
				},
			},
		},
	}

	t.Run("Numeric port string", func(t *testing.T) {
		svc := &corev1.Service{}
		port, err := resolveServicePort(svc, pod, "8080")
		assert.NoError(t, err)
		assert.Equal(t, int32(8080), port)
	})

	t.Run("Service port found with int target port", func(t *testing.T) {
		svc := &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		}
		port, err := resolveServicePort(svc, pod, "http")
		assert.NoError(t, err)
		assert.Equal(t, int32(8080), port)
	})

	t.Run("Service port found with string target port resolved via pod", func(t *testing.T) {
		svc := &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "metrics-port",
						TargetPort: intstr.FromString("metrics"),
					},
				},
			},
		}
		port, err := resolveServicePort(svc, pod, "metrics-port")
		assert.NoError(t, err)
		assert.Equal(t, int32(9100), port)
	})

	t.Run("Service port not found, falls back to pod port resolution", func(t *testing.T) {
		svc := &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{},
			},
		}
		port, err := resolveServicePort(svc, pod, "metrics")
		assert.NoError(t, err)
		assert.Equal(t, int32(9100), port)
	})

	t.Run("Service port not found and pod port not found", func(t *testing.T) {
		svc := &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{},
			},
		}
		_, err := resolveServicePort(svc, pod, "notexist")
		assert.Error(t, err)
	})
}

// NewPantheonClient

func TestNewPantheonClient(t *testing.T) {
	t.Run("Adds http prefix when missing", func(t *testing.T) {
		c := NewPantheonClient("localhost:8899", "prod", "token123")
		assert.Equal(t, "http://localhost:8899", c.ServerURL)
		assert.Equal(t, "prod", c.ClusterName)
		assert.Equal(t, "token123", c.AuthToken)
	})

	t.Run("Keeps https prefix", func(t *testing.T) {
		c := NewPantheonClient("https://myserver.example.com", "dev", "")
		assert.Equal(t, "https://myserver.example.com", c.ServerURL)
	})

	t.Run("Keeps http prefix and trims trailing slash", func(t *testing.T) {
		c := NewPantheonClient("http://myserver.example.com/", "dev", "")
		assert.Equal(t, "http://myserver.example.com", c.ServerURL)
	})
}

// PantheonClient HTTP methods

func TestFetchMonitorRules(t *testing.T) {
	expected := []model.MonitorRule{
		{ID: 1, Name: "rule-a", Type: "pod"},
	}

	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/ph/v1/monitors", r.URL.Path)
			assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expected)
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "tok")
		rules, err := c.FetchMonitorRules()
		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "rule-a", rules[0].Name)
	})

	t.Run("Non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "")
		_, err := c.FetchMonitorRules()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "not-json")
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "")
		_, err := c.FetchMonitorRules()
		assert.Error(t, err)
	})
}

func TestFetchRegisteredTargets(t *testing.T) {
	expected := []apitarget.TargetList{
		{ID: 42, Address: "10.0.0.1:9100", MetricPath: "/metrics"},
	}

	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expected)
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "mycluster", "")
		targets, err := c.FetchRegisteredTargets()
		assert.NoError(t, err)
		assert.Len(t, targets, 1)
		assert.Equal(t, uint(42), targets[0].ID)
	})

	t.Run("Non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "mycluster", "")
		_, err := c.FetchRegisteredTargets()
		assert.Error(t, err)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "{bad}")
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "mycluster", "")
		_, err := c.FetchRegisteredTargets()
		assert.Error(t, err)
	})
}

func TestRegisterTarget(t *testing.T) {
	item := apitarget.TargetItem{
		Address:    "10.0.0.5:9100",
		MetricPath: "/metrics",
		Labels:     map[string]string{"env": "prod"},
	}

	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "/ph/v1/targets", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "mytoken")
		err := c.RegisterTarget("rule-a", item)
		assert.NoError(t, err)
	})

	t.Run("Non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "bad request")
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "")
		err := c.RegisterTarget("rule-a", item)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")
	})
}

func TestDeleteTarget(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/ph/v1/targets/7", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "")
		err := c.DeleteTarget(7)
		assert.NoError(t, err)
	})

	t.Run("Non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "not found")
		}))
		defer srv.Close()

		c := NewPantheonClient(srv.URL, "c1", "tok")
		err := c.DeleteTarget(7)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")
	})
}

// PodReconciler helpers (deletePodTargets)

func TestDeletePodTargets(t *testing.T) {
	targets := []apitarget.TargetList{
		{
			ID:      1,
			Address: "10.0.0.1:9100",
			Labels: map[string]string{
				"kubernetes_namespace": "default",
				"kubernetes_pod_name":  "mypod",
			},
		},
		{
			ID:      2,
			Address: "10.0.0.2:9100",
			Labels: map[string]string{
				"kubernetes_namespace": "default",
				"kubernetes_pod_name":  "other-pod",
			},
		},
	}

	deletedIDs := []string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(targets)
		case r.Method == http.MethodDelete:
			deletedIDs = append(deletedIDs, r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	c := NewPantheonClient(srv.URL, "mycluster", "")
	r := &PodReconciler{PhClient: c}

	err := r.deletePodTargets("default", "mypod")
	assert.NoError(t, err)
	assert.Len(t, deletedIDs, 1)
	assert.Equal(t, "/ph/v1/targets/1", deletedIDs[0])
}

// ServiceReconciler helpers (deleteServiceTargets)

func TestDeleteServiceTargets(t *testing.T) {
	targets := []apitarget.TargetList{
		{
			ID:      10,
			Address: "10.0.0.3:9100",
			Labels: map[string]string{
				"kubernetes_namespace":    "prod",
				"kubernetes_service_name": "my-svc",
			},
		},
		{
			ID:      11,
			Address: "10.0.0.4:9100",
			Labels: map[string]string{
				"kubernetes_namespace":    "prod",
				"kubernetes_service_name": "other-svc",
			},
		},
	}

	deletedIDs := []string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(targets)
		case r.Method == http.MethodDelete:
			deletedIDs = append(deletedIDs, r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	c := NewPantheonClient(srv.URL, "mycluster", "")
	r := &ServiceReconciler{PhClient: c}

	err := r.deleteServiceTargets("prod", "my-svc")
	assert.NoError(t, err)
	assert.Len(t, deletedIDs, 1)
	assert.Equal(t, "/ph/v1/targets/10", deletedIDs[0])
}

// PodReconciler.Reconcile via fake k8s client

func newFakePod(namespace, name, podIP, nodeName string, labels map[string]string, ports []corev1.ContainerPort) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			Containers: []corev1.Container{
				{Ports: ports},
			},
		},
		Status: corev1.PodStatus{PodIP: podIP},
	}
}

func newFakeService(namespace, name string, labels, selector map[string]string, ports []corev1.ServicePort) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports:    ports,
		},
	}
}

// newTestScheme builds a scheme with core Kubernetes types registered.
func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	return s
}

// PodReconciler.Reconcile

func TestPodReconciler_Reconcile(t *testing.T) {
	scheme := newTestScheme()

	// Helper: build a fake Pantheon server returning given rules and targets.
	newSrv := func(rules []model.MonitorRule, targets []apitarget.TargetList, registerCalled *bool, deleteCalled *bool) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/ph/v1/monitors":
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(rules)
			case r.Method == http.MethodGet:
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(targets)
			case r.Method == http.MethodPut:
				if registerCalled != nil {
					*registerCalled = true
				}
				w.WriteHeader(http.StatusOK)
			case r.Method == http.MethodDelete:
				if deleteCalled != nil {
					*deleteCalled = true
				}
				w.WriteHeader(http.StatusOK)
			}
		}))
	}

	t.Run("Pod not found – deletes existing targets", func(t *testing.T) {
		// Pod doesn't exist in k8s, but 1 registered target exists.
		targets := []apitarget.TargetList{
			{ID: 1, Address: "10.0.0.1:9100", Labels: map[string]string{
				"kubernetes_namespace": "default", "kubernetes_pod_name": "mypod",
			}},
		}
		deleteCalled := false
		srv := newSrv(nil, targets, nil, &deleteCalled)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme) // no Pod added
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "mypod"},
		})
		assert.NoError(t, err)
		assert.True(t, deleteCalled)
	})

	t.Run("Pod has no matching rule – deletes existing targets", func(t *testing.T) {
		pod := newFakePod("default", "mypod", "10.0.0.1", "node1",
			map[string]string{"app": "something-else"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-a", Type: "pod", Namespace: "default", SelectorString: "app=portal", PortName: "http"},
		}
		targets := []apitarget.TargetList{
			{ID: 5, Address: "10.0.0.1:9100", Labels: map[string]string{
				"kubernetes_namespace": "default", "kubernetes_pod_name": "mypod",
			}},
		}
		deleteCalled := false
		srv := newSrv(rules, targets, nil, &deleteCalled)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "mypod"},
		})
		assert.NoError(t, err)
		assert.True(t, deleteCalled)
	})

	t.Run("Pod matches rule but has no IP – skip", func(t *testing.T) {
		pod := newFakePod("default", "mypod", "", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-a", Type: "pod", Namespace: "*", SelectorString: "app=portal", PortName: "http"},
		}
		srv := newSrv(rules, nil, nil, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "mypod"},
		})
		assert.NoError(t, err)
	})

	t.Run("Pod matches rule but port not found – skip", func(t *testing.T) {
		pod := newFakePod("default", "mypod", "10.0.0.1", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-a", Type: "pod", Namespace: "", SelectorString: "app=portal", PortName: "notexist"},
		}
		srv := newSrv(rules, nil, nil, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "mypod"},
		})
		assert.NoError(t, err)
	})

	t.Run("Pod matches rule – registers new target", func(t *testing.T) {
		pod := newFakePod("default", "mypod", "10.0.0.1", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-a", Type: "pod", Namespace: "default", SelectorString: "app=portal",
				PortName: "http", MetricPath: "/metrics", LabelsString: "team=sre", DropMetrics: "go_gc"},
		}
		registerCalled := false
		srv := newSrv(rules, []apitarget.TargetList{}, &registerCalled, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "mypod"},
		})
		assert.NoError(t, err)
		assert.True(t, registerCalled)
	})

	t.Run("Pod matches rule – already registered, no re-register", func(t *testing.T) {
		pod := newFakePod("default", "mypod", "10.0.0.1", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-a", Type: "pod", Namespace: "default", SelectorString: "app=portal",
				PortName: "http", MetricPath: "/metrics"},
		}
		// Already registered with matching labels
		targets := []apitarget.TargetList{
			{
				ID: 1, Address: "10.0.0.1:9100", MetricPath: "/metrics",
				Labels: map[string]string{
					"kubernetes_namespace": "default",
					"kubernetes_pod_name":  "mypod",
					"kubernetes_node_name": "node1",
					"monitor_rule":         "rule-a",
				},
			},
		}
		registerCalled := false
		srv := newSrv(rules, targets, &registerCalled, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "mypod"},
		})
		assert.NoError(t, err)
		assert.False(t, registerCalled)
	})

	t.Run("Pod rule namespace filter mismatch – no match", func(t *testing.T) {
		pod := newFakePod("other-ns", "mypod", "10.0.0.1", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-a", Type: "pod", Namespace: "default", SelectorString: "app=portal", PortName: "http"},
		}
		srv := newSrv(rules, []apitarget.TargetList{}, nil, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &PodReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "other-ns", Name: "mypod"},
		})
		assert.NoError(t, err)
	})
}

// ServiceReconciler.Reconcile

func TestServiceReconciler_Reconcile(t *testing.T) {
	scheme := newTestScheme()

	newSrv := func(rules []model.MonitorRule, targets []apitarget.TargetList, registerCalled *bool, deleteCalled *bool) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/ph/v1/monitors":
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(rules)
			case r.Method == http.MethodGet:
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(targets)
			case r.Method == http.MethodPut:
				if registerCalled != nil {
					*registerCalled = true
				}
				w.WriteHeader(http.StatusOK)
			case r.Method == http.MethodDelete:
				if deleteCalled != nil {
					*deleteCalled = true
				}
				w.WriteHeader(http.StatusOK)
			}
		}))
	}

	t.Run("Service not found – deletes existing targets", func(t *testing.T) {
		targets := []apitarget.TargetList{
			{ID: 20, Address: "10.0.0.5:9100", Labels: map[string]string{
				"kubernetes_namespace": "default", "kubernetes_service_name": "my-svc",
			}},
		}
		deleteCalled := false
		srv := newSrv(nil, targets, nil, &deleteCalled)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &ServiceReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-svc"},
		})
		assert.NoError(t, err)
		assert.True(t, deleteCalled)
	})

	t.Run("Service has no matching rule – deletes existing targets", func(t *testing.T) {
		svc := newFakeService("default", "my-svc",
			map[string]string{"app": "something-else"},
			map[string]string{"app": "something-else"},
			nil)

		rules := []model.MonitorRule{
			{Name: "rule-svc", Type: "service", Namespace: "default", SelectorString: "app=portal", PortName: "http"},
		}
		targets := []apitarget.TargetList{
			{ID: 21, Address: "10.0.0.5:9100", Labels: map[string]string{
				"kubernetes_namespace": "default", "kubernetes_service_name": "my-svc",
			}},
		}
		deleteCalled := false
		srv := newSrv(rules, targets, nil, &deleteCalled)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, svc)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &ServiceReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-svc"},
		})
		assert.NoError(t, err)
		assert.True(t, deleteCalled)
	})

	t.Run("Service matches rule but has no selector – skip", func(t *testing.T) {
		svc := newFakeService("default", "my-svc",
			map[string]string{"app": "portal"},
			nil, // empty selector
			nil)

		rules := []model.MonitorRule{
			{Name: "rule-svc", Type: "service", Namespace: "*", SelectorString: "app=portal", PortName: "8080"},
		}
		srv := newSrv(rules, []apitarget.TargetList{}, nil, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, svc)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &ServiceReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-svc"},
		})
		assert.NoError(t, err)
	})

	t.Run("Service matches rule, pod has no IP – skip pod", func(t *testing.T) {
		svc := newFakeService("default", "my-svc",
			map[string]string{"app": "portal"},
			map[string]string{"app": "portal"},
			[]corev1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(9100)}})

		pod := newFakePod("default", "mypod", "", "node1", // empty IP
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-svc", Type: "service", Namespace: "", SelectorString: "app=portal", PortName: "http"},
		}
		registerCalled := false
		srv := newSrv(rules, []apitarget.TargetList{}, &registerCalled, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, svc, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &ServiceReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-svc"},
		})
		assert.NoError(t, err)
		assert.False(t, registerCalled)
	})

	t.Run("Service matches rule – registers new target and cleans obsolete", func(t *testing.T) {
		svc := newFakeService("default", "my-svc",
			map[string]string{"app": "portal"},
			map[string]string{"app": "portal"},
			[]corev1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(9100)}})

		pod := newFakePod("default", "mypod", "10.0.0.9", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-svc", Type: "service", Namespace: "default", SelectorString: "app=portal",
				PortName: "http", MetricPath: "/metrics", LabelsString: "env=prod", DropMetrics: "go_gc"},
		}
		// An obsolete registered target (old pod IP not in discovered)
		targets := []apitarget.TargetList{
			{
				ID: 30, Address: "10.0.0.99:9100", MetricPath: "/metrics",
				Labels: map[string]string{
					"kubernetes_namespace": "default", "kubernetes_service_name": "my-svc",
				},
			},
		}
		registerCalled := false
		deleteCalled := false
		srv := newSrv(rules, targets, &registerCalled, &deleteCalled)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, svc, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &ServiceReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-svc"},
		})
		assert.NoError(t, err)
		assert.True(t, deleteCalled, "obsolete target should be deleted")
		assert.True(t, registerCalled, "new target should be registered")
	})

	t.Run("Service matches rule – already registered, no re-register", func(t *testing.T) {
		svc := newFakeService("default", "my-svc",
			map[string]string{"app": "portal"},
			map[string]string{"app": "portal"},
			[]corev1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(9100)}})

		pod := newFakePod("default", "mypod", "10.0.0.9", "node1",
			map[string]string{"app": "portal"},
			[]corev1.ContainerPort{{Name: "http", ContainerPort: 9100}})

		rules := []model.MonitorRule{
			{Name: "rule-svc", Type: "service", Namespace: "default", SelectorString: "app=portal",
				PortName: "http", MetricPath: "/metrics"},
		}
		targets := []apitarget.TargetList{
			{
				ID: 31, Address: "10.0.0.9:9100", MetricPath: "/metrics",
				Labels: map[string]string{
					"kubernetes_namespace":    "default",
					"kubernetes_service_name": "my-svc",
					"kubernetes_pod_name":     "mypod",
					"kubernetes_node_name":    "node1",
					"monitor_rule":            "rule-svc",
				},
			},
		}
		registerCalled := false
		srv := newSrv(rules, targets, &registerCalled, nil)
		defer srv.Close()

		c := fake.NewFakeClientWithScheme(scheme, svc, pod)
		ph := NewPantheonClient(srv.URL, "c1", "")
		r := &ServiceReconciler{Client: c, Scheme: scheme, PhClient: ph}

		_, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "my-svc"},
		})
		assert.NoError(t, err)
		assert.False(t, registerCalled)
	})
}
