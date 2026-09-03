package k8smodule

import (
	"log/slog"
	"os"
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ClientManager manages connection to the Kubernetes cluster.
type ClientManager struct {
	Clientset     *kubernetes.Clientset
	DynamicClient dynamic.Interface
	RESTMapper    *restmapper.DeferredDiscoveryRESTMapper
	Config        *rest.Config
	Endpoint      string
	Connected     bool
	Logger        *slog.Logger
}

// NewClientManager attempts to build a Kubernetes clientset using local kubeconfig
// or in-cluster config. If connection fails, it degrades gracefully.
func NewClientManager(logger *slog.Logger) *ClientManager {
	cm := &ClientManager{
		Logger: logger,
	}

	cfg, err := buildKubeConfig()
	if err != nil {
		logger.Warn("could not build kubeconfig, running in demo/offline mode", "error", err.Error())
		return cm
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Warn("failed to create kubernetes clientset", "error", err.Error())
		return cm
	}

	cm.Clientset = cs
	cm.Config = cfg
	cm.Endpoint = cfg.Host
	cm.Connected = true

	// Initialize dynamic client for arbitrary resource apply
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		logger.Warn("failed to create dynamic client", "error", err.Error())
	} else {
		cm.DynamicClient = dyn
	}

	// Initialize discovery client and RESTMapper for GVK to GVR translation
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		logger.Warn("failed to create discovery client", "error", err.Error())
	} else {
		cm.RESTMapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disc))
	}

	logger.Info("connected to kubernetes cluster", "host", cfg.Host)
	return cm
}

func buildKubeConfig() (*rest.Config, error) {
	// 1. Check KUBECONFIG env var
	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		if cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigEnv); err == nil {
			return cfg, nil
		}
	}

	// 2. Check ~/.kube/config
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			if cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath); err == nil {
				return cfg, nil
			}
		}
	}

	// 3. Fallback to in-cluster config (when running inside a Pod)
	return rest.InClusterConfig()
}
