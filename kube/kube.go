package kube

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/airware/vili/config"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/log"
)

// ExitingChan is a flag indicating that the server is exiting
var ExitingChan = make(chan struct{})

var kubeconfig *Config
var defaultClient *client

// Config is the kubernetes configuration
type Config struct {
	EnvConfigs map[string]*EnvConfig
}

// EnvConfig is an environment's kubernetes configuration
type EnvConfig struct {
	URL          string
	Namespace    string
	Token        string
	ClientCert   string
	ClientCACert string
	ClientKey    string

	client *client
}

// Create uses `kubectl create` to create the objects defined by `spec`
func Create(spec string) (map[string][]string, error) {
	out, err := kubectl(bytes.NewReader([]byte(spec)), "create", "-f", "-", "-o", "name")
	if err != nil {
		return nil, err
	}
	resources := make(map[string][]string)
	for _, resource := range strings.Fields(out) {
		parts := strings.SplitN(resource, "/", 2)
		resources[parts[0]] = append(resources[parts[0]], parts[1])
	}
	return resources, err
}

// Delete uses `kubectl delete` to delete the objects defined by `spec`
func Delete(spec string) (map[string][]string, error) {
	out, err := kubectl(bytes.NewReader([]byte(spec)), "delete", "-f", "-", "-o", "name")
	resources := make(map[string][]string)
	for _, resource := range strings.Fields(out) {
		parts := strings.SplitN(resource, "/", 2)
		resources[parts[0]] = append(resources[parts[0]], parts[1])
	}
	return resources, err
}

func kubectl(stdin io.Reader, args ...string) (string, error) {
	envConfig, ok := kubeconfig.EnvConfigs[config.GetString(config.DefaultEnv)]
	if !ok {
		token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err != nil {
			return "", err
		}
		envConfig = &EnvConfig{
			URL:   "https://kubernetes.default.svc.cluster.local",
			Token: string(token),
		}
	}
	kubeArgs := []string{"--server", envConfig.URL}
	if envConfig.Token != "" {
		kubeArgs = append(kubeArgs, "--token", envConfig.Token)
	}
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"); !os.IsNotExist(err) {
		kubeArgs = append(kubeArgs, "--certificate-authority", "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	}
	kubeArgs = append(kubeArgs, args...)

	cmd := exec.Command("kubectl", kubeArgs...)
	cmd.Stdin = stdin

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func getClient(env string) (*client, error) {
	envConfig, ok := kubeconfig.EnvConfigs[env]
	if ok {
		return envConfig.client, nil
	}
	baseClient, err := getDefaultClient()
	if err != nil {
		return nil, err
	}
	c := *baseClient
	c.namespace = env
	return &c, nil
}

func getDefaultClient() (*client, error) {
	if defaultClient == nil {
		defaultEnv := config.GetString(config.DefaultEnv)
		if c, ok := kubeconfig.EnvConfigs[defaultEnv]; ok {
			defaultClient = c.client
			return defaultClient, nil
		}
		token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err != nil {
			return nil, err
		}

		caCert, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		defaultClient = &client{
			httpClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: caCertPool,
					},
				},
				Timeout: 5 * time.Second,
			},
			httpClientNoTimeout: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: caCertPool,
					},
				},
			},
			url:   "https://kubernetes.default.svc.cluster.local",
			token: string(token),
		}
	}
	return defaultClient, nil
}

// Init initializes the kubernetes service with the given config
func Init(c *Config) error {
	kubeconfig = c
	for env, envConfig := range kubeconfig.EnvConfigs {
		var tr *http.Transport
		if envConfig.URL == "" {
			envConfig.URL = "https://kubernetes.default.svc.cluster.local"
			token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
			if err != nil {
				return err
			}
			envConfig.Token = string(token)

			caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
			if err != nil {
				return err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: caCertPool},
			}
		} else {
			tr = &http.Transport{}
		}

		if envConfig.Namespace == "" {
			envConfig.Namespace = env
		}

		if envConfig.ClientCert != "" {
			cert, err := tls.LoadX509KeyPair(envConfig.ClientCert, envConfig.ClientKey)
			if err != nil {
				return err
			}
			caCert, err := ioutil.ReadFile(envConfig.ClientCACert)
			if err != nil {
				return err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tr.TLSClientConfig = &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			}
		}

		envConfig.client = &client{
			httpClient: &http.Client{
				Transport: tr,
				Timeout:   5 * time.Second,
			},
			httpClientNoTimeout: &http.Client{
				Transport: tr,
			},
			url:       envConfig.URL,
			namespace: envConfig.Namespace,
			token:     envConfig.Token,
		}
	}
	return nil
}

type client struct {
	httpClient          *http.Client
	httpClientNoTimeout *http.Client
	url                 string
	token               string
	namespace           string
}

func (c *client) createRequest(method, path string, query *url.Values, body io.Reader, watch bool) (*http.Request, error) {
	// get url string
	apiBase := fmt.Sprintf("%s/api/v1/", c.url)
	if strings.HasPrefix(path, "deployments") ||
		strings.HasPrefix(path, "replicasets") {
		apiBase = fmt.Sprintf("%s/apis/extensions/v1beta1/", c.url)
	}
	if strings.HasPrefix(path, "jobs") {
		apiBase = fmt.Sprintf("%s/apis/batch/v1/", c.url)
	}
	if !strings.HasPrefix(path, "namespace") && !strings.HasPrefix(path, "node") {
		path = fmt.Sprintf("namespaces/%s/%s", c.namespace, path)
	}
	if watch && !strings.HasSuffix(path, "/log") {
		path = "watch/" + path
	}
	urlStr := apiBase + path
	if query != nil {
		urlStr += "?" + query.Encode()
	}
	log.WithField("method", method).WithField("url", urlStr).Debug("making kubernetes request")

	// create request
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if method == "PATCH" {
		req.Header.Add("Content-Type", "application/merge-patch+json")
	}
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	return req, nil
}

func (c *client) getRequestBytes(method, path string, query *url.Values, body io.Reader) (respBody []byte, respStatus *unversioned.Status, err error) {
	// create request
	req, err := c.createRequest(method, path, query, body, false)
	if err != nil {
		return
	}

	// send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// read response body
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// parse status
	if resp.Header.Get("Content-Type") == "application/json" {
		typeMeta := new(unversioned.TypeMeta)
		err = json.Unmarshal(respBody, typeMeta)
		if err != nil {
			return
		}
		if typeMeta.Kind == "Status" {
			respStatus = new(unversioned.Status)
			err = json.Unmarshal(respBody, respStatus)
			return
		}
	}
	return
}

func (c *client) unmarshalRequest(method, path string, query *url.Values, body io.Reader, dest interface{}) (respStatus *unversioned.Status, err error) {
	respBody, respStatus, err := c.getRequestBytes(method, path, query, body)
	if respStatus != nil || err != nil {
		return
	}
	if dest != nil {
		err = json.Unmarshal(respBody, dest)
	}
	return
}

func invalidEnvError(env string) error {
	return errors.BadRequest(fmt.Sprintf("Invalid environment %s", env))
}
