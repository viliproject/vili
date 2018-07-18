package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/repository"
	"github.com/airware/vili/session"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/types"
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func releasesGetHandler(c echo.Context) error {
	env := c.Param("env")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	releaseEnvName := env
	if environment.DeployedToEnv != "" {
		releaseEnvName = environment.DeployedToEnv
	}

	websocket.Handler(func(ws *websocket.Conn) {
		err = releasesWatchHandler(ws, releaseEnvName)
		ws.Close()
	}).ServeHTTP(c.Response(), c.Request())
	return err
}

func releasesWatchHandler(ws *websocket.Conn, env string) error {
	eventsChan := make(chan firebase.Event)
	stopped := false
	var waitGroup sync.WaitGroup

	go func() {
		var cmd interface{}
		err := websocket.JSON.Receive(ws, cmd)
		if err == io.EOF {
			stopped = true
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for event := range eventsChan {
			// Parse event into a release event
			releaseEvent := getReleaseEvent(event)
			if releaseEvent != nil {
				err := websocket.JSON.Send(ws, releaseEvent)
				if err != nil {
					log.WithError(err).Warn("error writing to websocket stream")
				}
			}
		}
	}()

	err := firebase.Watch(fmt.Sprintf("/releases/%s", env), eventsChan)
	close(eventsChan)
	waitGroup.Wait()
	if !stopped {
		websocket.JSON.Send(ws, webSocketCloseMessage)
	}
	return err
}

// ReleaseEvent represents a change to a release
type ReleaseEvent struct {
	Type   string           `json:"type"`
	Object *types.Release   `json:"object"`
	List   []*types.Release `json:"list"`
}

func getReleaseEvent(firebaseEvent firebase.Event) *ReleaseEvent {
	switch firebaseEvent.Type {
	case "put":
		data, _ := json.Marshal(firebaseEvent.Data)
		if firebaseEvent.Path == "/" {
			releases := map[string]*types.Release{}
			err := json.Unmarshal(data, &releases)
			if err != nil {
				log.WithError(err).Warn("error parsing releases json")
				return nil
			}
			releaseList := []*types.Release{}
			for _, release := range releases {
				if len(release.Rollouts) == 0 {
					release.Rollouts = []*types.ReleaseRollout{}
				}
				releaseList = append(releaseList, release)
			}
			return &ReleaseEvent{
				Type: "INIT",
				List: releaseList,
			}
		}
		pathSlice := strings.Split(strings.TrimPrefix(firebaseEvent.Path, "/"), "/")
		if len(pathSlice) != 1 {
			return nil
		}
		if firebaseEvent.Data == nil {
			return &ReleaseEvent{
				Type: "DELETED",
				Object: &types.Release{
					Name: pathSlice[0],
				},
			}
		}
		release := &types.Release{
			Name:     pathSlice[0],
			Rollouts: []*types.ReleaseRollout{},
		}
		err := json.Unmarshal(data, release)
		if err != nil {
			log.WithError(err).Warn("error parsing release json")
			return nil
		}
		return &ReleaseEvent{
			Type:   "MODIFIED",
			Object: release,
		}
	}
	return nil
}

func releaseSpecGetHandler(c echo.Context) error {
	env := c.Param("env")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	release := new(types.Release)
	spec, err := templates.Release(environment.Name, environment.Branch)
	if err != nil {
		return err
	}
	err = spec.Parse(release)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, release)
}

func releaseCreateHandler(c echo.Context) error {
	env := c.Param("env")
	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	release := new(types.Release)
	if c.QueryParam("latest") == "" {
		// get release metadata from the request
		err = json.NewDecoder(c.Request().Body).Decode(release)
		if err != nil {
			return err
		}
		// check required fields
		if release.Name == "" {
			return errors.BadRequest("Release name required")
		}
	} else {
		// get spec for this environment
		spec, err := templates.Release(environment.Name, environment.Branch)
		if err != nil {
			return err
		}
		if err = spec.Parse(release); err != nil {
			return err
		}
		release.Name = time.Now().Format("latest-20060102150405")
		// get waves by selecting the latest version of each job and app
		if populateReleaseLatestVersions(c.Request().Context(), environment, release) {
			return errors.InternalServerError()
		}
	}
	// validate fields
	if release.Link != "" && !govalidator.IsURL(release.Link) {
		release.Link = ""
	}
	// set hardcoded fields
	release.TargetEnv = environment.DeployedToEnv
	if release.TargetEnv == "" {
		release.TargetEnv = env
	}
	release.CreatedAt = time.Now()
	release.CreatedBy = c.Get("user").(*session.User).Username

	// check for conflicts
	existingRelease, err := getReleaseValue(release.TargetEnv, release.Name)
	if err != nil {
		return err
	}
	if existingRelease.Name != "" {
		return errors.Conflict("Release already exists")
	}

	// save release to the database
	err = setReleaseValue(release)
	if err != nil {
		return err
	}
	// send notifications
	slackMessage := fmt.Sprintf("release *%s* created by *%s*", release.Name, release.CreatedBy)
	if release.Link != "" {
		slackMessage += fmt.Sprintf(" - <%s|release notes>", release.Link)
	}
	err = slack.PostLogMessage(slackMessage, log.InfoLevel)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, release)
}

func releaseDeleteHandler(c echo.Context) error {
	env := c.Param("env")
	name := c.Param("release")
	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	if environment.DeployedToEnv == "" {
		return errors.BadRequest("Cannot create release from a non-approval environment")
	}

	err = deleteRelease(environment.DeployedToEnv, name)
	if err != nil {
		return err
	}
	slackMessage := fmt.Sprintf("release *%s* deleted by *%s*", name, c.Get("user").(*session.User).Username)
	err = slack.PostLogMessage(slackMessage, log.InfoLevel)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func releaseDeployHandler(c echo.Context) error {
	env := c.Param("env")
	name := c.Param("release")
	environment, err := environments.Get(env)
	if err != nil {
		return err
	}
	releaseEnv := environment.DeployedToEnv
	if releaseEnv == "" {
		releaseEnv = env
	}
	release, err := getReleaseValue(releaseEnv, name)
	if err != nil {
		return err
	}
	if release.Name == "" {
		return errors.NotFound("Release not found")
	}
	// create release rollout
	releaseRollout, err := createReleaseRollout(release, env, c.Get("user").(*session.User).Username)
	if err != nil {
		return err
	}
	// deploy release
	go func() {
		err = deployRelease(c.Request().Context(), release, releaseRollout)
		if err != nil {
			log.WithError(err).Error("failed release rollout")
		}
	}()
	return c.JSON(http.StatusOK, releaseRollout)
}

func populateReleaseLatestVersions(ctx context.Context, environment *environments.Environment, release *types.Release) (failed bool) {
	var wg sync.WaitGroup
	for _, wave := range release.Waves {
		for _, target := range wave.Targets {
			wg.Add(1)
			go func(target *types.ReleaseTarget) {
				if err := populateReleaseTargetLatestVersion(ctx, environment, target); err != nil {
					log.WithError(err).Error("failed populating latest versions for release target")
					failed = true
				}
				wg.Done()
			}(target)
		}
	}
	wg.Wait()
	return
}

func populateReleaseTargetLatestVersion(ctx context.Context, environment *environments.Environment, target *types.ReleaseTarget) error {
	switch target.Type {
	case types.ReleaseTargetTypeAction:
		target.Branch = environment.Branch
		return nil
	case types.ReleaseTargetTypeApp:
		imageRepo, err := getDeploymentImageRepo(target.Name, environment.Name, environment.RepositoryBranches[0])
		if err != nil {
			return err
		}
		images, err := repository.GetDockerRepository(ctx, imageRepo, environment.RepositoryBranches)
		if err != nil {
			return err
		}
		if len(images) == 0 {
			return fmt.Errorf("Target %s does not have any images in the repository", target.Name)
		}
		image := images[0]
		target.Tag = image.Tag
		target.Branch = image.Branch
		return nil
	case types.ReleaseTargetTypeJob:
		imageRepo, err := getJobImageRepo(target.Name, environment.Name, environment.RepositoryBranches[0])
		if err != nil {
			return err
		}
		images, err := repository.GetDockerRepository(ctx, imageRepo, environment.RepositoryBranches)
		if err != nil {
			return err
		}
		if len(images) == 0 {
			return fmt.Errorf("Target %s does not have any images in the repository", target.Name)
		}
		image := images[0]
		target.Tag = image.Tag
		target.Branch = image.Branch
		return nil
	}
	return nil
}

func createReleaseRollout(release *types.Release, env, username string) (*types.ReleaseRollout, error) {
	releaseRollout := &types.ReleaseRollout{
		ID:        len(release.Rollouts) + 1,
		Env:       env,
		RolloutAt: time.Now(),
		RolloutBy: username,
		Status:    types.RolloutStatusDeploying,
	}
	release.Rollouts = append(release.Rollouts, releaseRollout)
	// set status of all waves to "new"
	for range release.Waves {
		releaseRolloutWave := &types.ReleaseRolloutWave{
			Status: types.RolloutStatusNew,
		}
		releaseRollout.Waves = append(releaseRollout.Waves, releaseRolloutWave)
	}
	return releaseRollout, setReleaseValue(release)
}

func deployRelease(ctx context.Context, release *types.Release, releaseRollout *types.ReleaseRollout) error {
	// deploy each wave in order
	for ix, wave := range release.Waves {
		// set status to deploying
		releaseRolloutWave := releaseRollout.Waves[ix]
		releaseRolloutWave.Status = types.RolloutStatusDeploying
		if err := setReleaseValue(release); err != nil {
			return err
		}
		// deploy
		if deployReleaseWave(ctx, wave, releaseRollout) {
			releaseRollout.Status = types.RolloutStatusFailed
			releaseRolloutWave.Status = types.RolloutStatusFailed
			return setReleaseValue(release)
		}
		// set status to deployed
		releaseRolloutWave.Status = types.RolloutStatusDeployed
		if err := setReleaseValue(release); err != nil {
			return err
		}
	}
	releaseRollout.Status = types.RolloutStatusDeployed
	return setReleaseValue(release)
}

func deployReleaseWave(ctx context.Context, wave *types.ReleaseWave, releaseRollout *types.ReleaseRollout) bool {
	log.Debugf("Deploying wave with %d targets", len(wave.Targets))
	// deploy targets in parallel
	var wg sync.WaitGroup
	var failed bool
	for _, target := range wave.Targets {
		wg.Add(1)
		go func(target *types.ReleaseTarget) {
			defer wg.Done()
			if err := deployReleaseTarget(ctx, target, releaseRollout); err != nil {
				log.WithError(err).Error("failed deploying target")
				failed = true
			}
		}(target)
	}
	wg.Wait()
	log.Debugf("Deployed wave with %d targets", len(wave.Targets))
	return failed
}

func deployReleaseTarget(ctx context.Context, target *types.ReleaseTarget, releaseRollout *types.ReleaseRollout) error {
	switch target.Type {
	case types.ReleaseTargetTypeAction:
		log.Debugf(
			"Executing action %s, from branch %s to env %s, requested by %s",
			target.Name, target.Branch, releaseRollout.Env, releaseRollout.RolloutBy)
		switch target.Name {
		case "syncConfigMaps":
			return syncConfigMaps(target, releaseRollout)
		}
	case types.ReleaseTargetTypeApp:
		log.Debugf(
			"Rolling out deployment %s, tag %s to env %s, requested by %s",
			target.Name, target.Tag, releaseRollout.Env, releaseRollout.RolloutBy)
		rollout := &Rollout{
			Env:            releaseRollout.Env,
			Username:       releaseRollout.RolloutBy,
			DeploymentName: target.Name,
			Branch:         target.Branch,
			Tag:            target.Tag,
		}
		return rollout.Run(ctx, false)
	case types.ReleaseTargetTypeJob:
		log.Debugf(
			"Running job %s, tag %s to env %s, requested by %s",
			target.Name, target.Tag, releaseRollout.Env, releaseRollout.RolloutBy)
		jobRun := &JobRun{
			Env:      releaseRollout.Env,
			Username: releaseRollout.RolloutBy,
			JobName:  target.Name,
			Branch:   target.Branch,
			Tag:      target.Tag,
		}
		return jobRun.Run(ctx, false)
	}
	return nil
}

func syncConfigMaps(target *types.ReleaseTarget, releaseRollout *types.ReleaseRollout) error {
	configmapNames, err := templates.ConfigMaps(releaseRollout.Env, target.Branch)
	if err != nil {
		return err
	}
	endpoint := kube.GetClient(releaseRollout.Env).ConfigMaps()
	for _, configmapName := range configmapNames {
		configmapTemplate, err := templates.ConfigMap(releaseRollout.Env, target.Branch, configmapName)
		if err != nil {
			return err
		}
		configmap := new(corev1.ConfigMap)
		err = configmapTemplate.Parse(configmap)
		if err != nil {
			return err
		}
		existingConfigMap, err := endpoint.Get(configmapName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if existingConfigMap == nil {
			_, err = endpoint.Create(configmap)
		} else {
			_, err = endpoint.Update(configmap)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func getReleaseValue(env, name string) (*types.Release, error) {
	release := new(types.Release)
	return release, firebase.Database().Child("releases").Child(env).Child(name).Value(release)
}

func setReleaseValue(release *types.Release) error {
	return firebase.Database().Child("releases").Child(release.TargetEnv).Child(release.Name).Set(release)
}

func deleteRelease(env, name string) error {
	return firebase.Database().Child("releases").Child(env).Child(name).Remove()
}
