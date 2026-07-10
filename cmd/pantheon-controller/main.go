package main

import (
	"flag"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cylonchau/pantheon/pkg/controller"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func main() {
	fs := flag.NewFlagSet("pantheon-controller", flag.ExitOnError)
	klog.InitFlags(fs)
	ctrl.SetLogger(klog.NewKlogr())
	defer klog.Flush()

	serverURL := fs.String("pantheon-server", "http://localhost:8899", "Pantheon server endpoint URL.")
	clusterName := fs.String("cluster-name", "k8s-prod-china", "Unique identifier name for this Kubernetes cluster.")
	authToken := fs.String("auth-token", "", "Optional authorization token for accessing Pantheon server APIs.")
	kubeconfig := fs.String("kubeconfig", "", "Absolute path to the kubeconfig file (optional, only used when running outside cluster).")
	syncInterval := fs.Duration("sync-interval", 30*time.Second, "Synchronization period for the controller manager to force full resync of watched resources.")

	_ = fs.Parse(os.Args[1:])

	if *serverURL == "" {
		klog.Fatal("flag --pantheon-server is required")
	}

	// 1. Build Kubernetes REST Config
	var config *rest.Config
	var err error

	if *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			klog.Fatalf("Error building kubeconfig from path %q: %v", *kubeconfig, err)
		}
	} else {
		klog.Info("No kubeconfig provided. Attempting to use in-cluster config...")
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("Error building in-cluster REST config: %v", err)
		}
	}

	// 2. Create the Manager
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:             scheme,
		LeaderElection:     false,
		MetricsBindAddress: "0",
		SyncPeriod:         syncInterval,
	})
	if err != nil {
		klog.Fatalf("Unable to start manager: %v", err)
	}

	// 3. Initialize PantheonClient
	phClient := controller.NewPantheonClient(*serverURL, *clusterName, *authToken)

	// 4. Register Pod Reconciler
	podReconciler := &controller.PodReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		PhClient: phClient,
	}
	err = ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(podReconciler)
	if err != nil {
		klog.Fatalf("Unable to create Pod controller: %v", err)
	}

	// 5. Register Service Reconciler
	svcReconciler := &controller.ServiceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		PhClient: phClient,
	}
	err = ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(svcReconciler)
	if err != nil {
		klog.Fatalf("Unable to create Service controller: %v", err)
	}

	// 6. Start the Manager
	klog.Infof("Starting Pantheon Kubernetes Operator for cluster %q", *clusterName)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatalf("Problem running manager: %v", err)
	}
}
