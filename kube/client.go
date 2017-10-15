package kube

import (
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ExitingChan is a flag indicating that the server is exiting
var ExitingChan = make(chan struct{})

var config *Config
var defaultRestConfig *rest.Config
var defaultClient client.Interface

// Config is the kubernetes configuration
type Config struct {
	EnvConfigs            map[string]*EnvConfig
	DefaultKubeConfigPath string
}

// EnvConfig is an environment's kubernetes configuration
type EnvConfig struct {
	Namespace      string
	KubeConfigPath string
	url            string
	token          string

	client *Client
}

// Client is just a basic wrapper around the unversioned client with helper methods
type Client struct {
	client.Interface
	namespace string
}

// Init initializes the kubernetes service with the given config
func Init(c *Config) error {
	config = c
	// get the default client
	var err error
	defaultRestConfig, err = newConfig(c.DefaultKubeConfigPath)
	if err != nil {
		return err
	}
	defaultClient, err = client.NewForConfig(defaultRestConfig)
	if err != nil {
		return err
	}

	// get the env clients
	for env, envConfig := range config.EnvConfigs {
		kubeConfig, err := newConfig(envConfig.KubeConfigPath)
		if err != nil {
			return err
		}
		kc, err := client.NewForConfig(kubeConfig)
		if err != nil {
			return err
		}
		namespace := envConfig.Namespace
		if namespace == "" {
			namespace = env
		}
		c := &Client{
			Interface: kc,
			namespace: namespace,
		}
		if err := c.Ping(); err != nil {
			return err
		}
		envConfig.client = c
	}
	return nil
}

// newConfig will either take the host string provided and return a config or attempt to find a
// reasonable config based on environment variables or DNS addresses expected in a k8s cluster.
func newConfig(kubeConfigPath string) (cfg *rest.Config, err error) {
	// Use kubeConfigPath if set
	if kubeConfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	} else {
		// Otherwise assume we're in a pod
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return
	}
	// Ignore checking the error since we're returning below anyway
	err = rest.LoadTLSFiles(cfg)
	return
}

// Ping checks the k8s /healthz endpoint and returns error if there's an error
func (k *Client) Ping() error {
	req := k.Core().RESTClient().Get()
	req.AbsPath("healthz", "ping")
	res := req.Do()
	return res.Error()
}

// GetClient returns the client for the given env
func GetClient(env string) *Client {
	if envConfig, ok := config.EnvConfigs[env]; ok {
		return envConfig.client
	}
	return &Client{
		Interface: defaultClient,
		namespace: env,
	}
}
