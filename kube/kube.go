package kube

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/airware/vili/kube/unversioned"
)

var config *Config

// Config is the kubernetes configuration
type Config struct {
	EnvConfigs map[string]*EnvConfig
}

// EnvConfig is an environment's kubernetes configuration
type EnvConfig struct {
	URL       string
	Namespace string
	Token     string
	Cert      string

	client *client
}

// Init initializes the kubernetes service with the given config
func Init(c *Config) error {
	config = c
	for env, envConfig := range config.EnvConfigs {
		var tr *http.Transport
		if envConfig.URL == "" {
			envConfig.URL = "https://kubernetes.default.svc.cluster.local"
			token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
			if err != nil {
				return err
			}
			envConfig.Token = string(token)

			caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
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

		envConfig.client = &client{
			httpClient: &http.Client{
				Transport: tr,
				Timeout:   5 * time.Second,
			},
			url:       envConfig.URL,
			namespace: envConfig.Namespace,
			token:     envConfig.Token,
		}
	}
	return nil
}

type client struct {
	httpClient *http.Client
	url        string
	token      string
	namespace  string
}

func (c *client) makeRequestRaw(method, path string, body io.Reader) ([]byte, *unversioned.Status, error) {
	if !strings.HasPrefix(path, "namespace") && !strings.HasPrefix(path, "node") {
		path = fmt.Sprintf("namespaces/%s/%s", c.namespace, path)
	}
	urlStr := fmt.Sprintf("%s/api/v1/%s", c.url, path)
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, nil, err
	}
	if method == "PATCH" {
		req.Header.Add("Content-Type", "application/merge-patch+json")
	}
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.Header.Get("Content-Type") == "application/json" {
		typeMeta := &unversioned.TypeMeta{}
		err = json.Unmarshal(respBody, typeMeta)
		if err != nil {
			return nil, nil, err
		}
		if typeMeta.Kind == "Status" {
			respStatus := &unversioned.Status{}
			err = json.Unmarshal(respBody, respStatus)
			if err != nil {
				return nil, nil, err
			}
			return nil, respStatus, nil
		}
	}
	return respBody, nil, nil
}

func (c *client) makeRequest(method, path string, body io.Reader, dest interface{}) (*unversioned.Status, error) {
	respBody, status, err := c.makeRequestRaw(method, path, body)
	if status != nil || err != nil {
		return status, err
	}
	return nil, json.Unmarshal(respBody, dest)
}

func invalidEnvError(env string) error {
	return fmt.Errorf("Invalid environment %s", env)
}
