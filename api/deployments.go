package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/repository"
	"github.com/airware/vili/server"
	"github.com/airware/vili/templates"
	"github.com/labstack/echo"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deploymentsGetHandler(c echo.Context) error {
	env := c.Param("env")

	endpoint := kube.GetClient(env).Deployments()
	query := getListOptionsFromRequest(c)

	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, query, endpoint.Watch)
	}

	resp, err := endpoint.List(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

type deploymentRepositoryResponse struct {
	Images []*repository.Image `json:"images,omitempty"`
}

func deploymentRepositoryGetHandler(c echo.Context) error {
	env := c.Param("env")
	deployment := c.Param("deployment")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(deploymentRepositoryResponse)
	images, err := repository.GetDockerRepository(deployment, environment.RepositoryBranches)
	if err != nil {
		return err
	}
	resp.Images = images

	return c.JSON(http.StatusOK, resp)
}

type deploymentSpecResponse struct {
	Spec string `json:"spec,omitempty"`
}

func deploymentSpecGetHandler(c echo.Context) error {
	env := c.Param("env")
	deployment := c.Param("deployment")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(deploymentSpecResponse)
	body, err := templates.Deployment(environment.Name, environment.Branch, deployment)
	if err != nil {
		return err
	}
	resp.Spec = string(body)

	return c.JSON(http.StatusOK, resp)
}

func deploymentServiceGetHandler(c echo.Context) error {
	env := c.Param("env")
	deployment := c.Param("deployment")

	endpoint := kube.GetClient(env).Services()

	service, err := endpoint.Get(deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, service)
}

func deploymentServiceCreateHandler(c echo.Context) error {
	env := c.Param("env")
	deploymentName := c.Param("deployment")

	endpoint := kube.GetClient(env).Services()

	failed := false

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	var deploymentTemplate templates.Template
	var waitGroup sync.WaitGroup

	// deploymentTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Deployment(environment.Name, environment.Branch, deploymentName)
		if err != nil {
			log.Error(err)
			failed = true
		}
		deploymentTemplate = body
	}()

	// service
	_, err = endpoint.Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		if statusError, ok := err.(*kubeErrors.StatusError); !ok || statusError.Status().Code != http.StatusNotFound {
			// only return error if the error is something other than NotFound
			return err
		}
	} else {
		return server.ErrorResponse(c, errors.Conflict("Service exists"))
	}

	waitGroup.Wait()

	deployment := &extv1beta1.Deployment{}
	err = deploymentTemplate.Parse(deployment)
	if err != nil {
		return err
	}

	deploymentPort, err := getPortFromDeployment(deployment)
	if err != nil {
		return err
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Protocol: "TCP",
					Port:     deploymentPort,
				},
			},
			Selector: map[string]string{
				"app": deploymentName,
			},
		},
	}

	resp, err := endpoint.Create(service)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

type deploymentActionRequest struct {
	Replicas   *int32 `json:"replicas"`
	ToRevision int64  `json:"toRevision"`
}

const (
	deploymentActionResume   = "resume"
	deploymentActionPause    = "pause"
	deploymentActionRollback = "rollback"
	deploymentActionScale    = "scale"
)

func deploymentActionHandler(c echo.Context) (err error) {
	env := c.Param("env")
	deploymentName := c.Param("deployment")
	action := c.Param("action")

	kubeClient := kube.GetClient(env)
	endpoint := kubeClient.Deployments()

	deployment, err := endpoint.Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	actionRequest := new(deploymentActionRequest)
	// ignore errors, as not all requests have a body
	json.NewDecoder(c.Request().Body).Decode(actionRequest)

	var resp interface{}

	switch action {
	case deploymentActionResume:
		deployment.Spec.Paused = false // TODO use Patch?
		resp, err = endpoint.Update(deployment)
	case deploymentActionPause:
		deployment.Spec.Paused = true
		resp, err = endpoint.Update(deployment)
	case deploymentActionRollback:
		err = endpoint.Rollback(&extv1beta1.DeploymentRollback{
			Name: deploymentName,
			RollbackTo: extv1beta1.RollbackConfig{
				Revision: actionRequest.ToRevision,
			},
		})
	case deploymentActionScale:
		if actionRequest.Replicas == nil {
			return server.ErrorResponse(c, errors.BadRequest("Replicas missing from scale request"))
		}
		resp, err = endpoint.UpdateScale(deploymentName, &extv1beta1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: kubeClient.Namespace(),
			},
			Spec: extv1beta1.ScaleSpec{
				Replicas: *actionRequest.Replicas,
			},
		})

	default:
		return server.ErrorResponse(c, errors.NotFound(fmt.Sprintf("Action %s not found", action)))
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func getRolloutHistoryForDeployment(env string, deployment *extv1beta1.Deployment) ([]*extv1beta1.ReplicaSet, error) {
	var selector []string
	for k, v := range deployment.Spec.Selector.MatchLabels {
		selector = append(selector, fmt.Sprintf("%s=%s", k, v))
	}
	replicaSetList, err := kube.GetClient(env).ReplicaSets().List(getListOptionsForDeployment(deployment))
	if err != nil {
		return nil, err
	}
	if replicaSetList == nil {
		return nil, fmt.Errorf("No replicaSet found for deployment %v", *deployment)
	}
	history := byRevision{}
	for _, replicaSet := range replicaSetList.Items {
		rs := replicaSet
		history = append(history, &rs)
	}
	sort.Sort(history)
	return history, nil
}

func getReplicaSetForDeployment(deployment *extv1beta1.Deployment, history []*extv1beta1.ReplicaSet) (*extv1beta1.ReplicaSet, error) {
	deploymentRevision := deployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	for _, replicaSet := range history {
		rsRevision := replicaSet.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if deploymentRevision == rsRevision {
			return replicaSet, nil
		}
	}
	return nil, fmt.Errorf("No replicaSet found for deployment %v", *deployment)
}

type byRevision []*extv1beta1.ReplicaSet

func (s byRevision) Len() int      { return len(s) }
func (s byRevision) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byRevision) Less(i, j int) bool {
	ri, _ := strconv.Atoi(s[i].ObjectMeta.Annotations["deployment.kubernetes.io/revision"])
	rj, _ := strconv.Atoi(s[j].ObjectMeta.Annotations["deployment.kubernetes.io/revision"])
	return ri > rj
}
