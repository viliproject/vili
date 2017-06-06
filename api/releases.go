package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/asaskevich/govalidator"
	echo "gopkg.in/labstack/echo.v1"

	"github.com/airware/vili/docker"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/session"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/types"
)

func releasesGetHandler(c *echo.Context) error {
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
	Type   string             `json:"type"`
	Object *types.Release     `json:"object"`
	List   *types.ReleaseList `json:"list"`
}

func getReleaseEvent(firebaseEvent firebase.Event) *ReleaseEvent {
	switch firebaseEvent.Type {
	case "put":
		data, _ := json.Marshal(firebaseEvent.Data)
		if firebaseEvent.Path == "/" {
			releases := map[string]types.Release{}
			err := json.Unmarshal(data, &releases)
			if err != nil {
				log.WithError(err).Warn("error parsing releases json")
				return nil
			}
			releaseList := &types.ReleaseList{
				Items: []types.Release{},
			}
			for _, release := range releases {
				releaseList.Items = append(releaseList.Items, release)
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
			Name: pathSlice[0],
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

func releaseSpecGetHandler(c *echo.Context) error {
	env := c.Param("env")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	release := new(types.Release)
	spec, err := templates.Release(environment.Name)
	if err != nil {
		return err
	}
	err = spec.Parse(release)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, release)
}

func releaseCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	release := new(types.Release)
	if c.Query("latest") == "" {
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
		spec, err := templates.Release(environment.Name)
		if err != nil {
			return err
		}
		if err = spec.Parse(release); err != nil {
			return err
		}
		release.Name = time.Now().Format("latest-20060102150405")
		// get waves by selecting the latest version of each job and app
		if populateReleaseLatestVersions(environment, release) {
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

func releaseDeleteHandler(c *echo.Context) error {
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

func releaseDeployHandler(c *echo.Context) error {
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
		err = deployRelease(release, releaseRollout)
		if err != nil {
			log.WithError(err).Error("failed release rollout")
		}
	}()
	return c.JSON(http.StatusOK, releaseRollout)
}

func populateReleaseLatestVersions(environment *environments.Environment, release *types.Release) (failed bool) {
	var wg sync.WaitGroup
	for _, wave := range release.Waves {
		for _, target := range wave.Targets {
			wg.Add(1)
			go func(target *types.ReleaseTarget) {
				if err := populateReleaseTargetLatestVersion(environment, target); err != nil {
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

func populateReleaseTargetLatestVersion(environment *environments.Environment, target *types.ReleaseTarget) error {
	switch target.Type {
	case types.ReleaseTargetTypeAction:
		target.Branch = environment.Branch
		return nil
	case types.ReleaseTargetTypeApp, types.ReleaseTargetTypeJob:
		images, err := docker.GetRepository(target.Name, environment.RepositoryBranches())
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

func deployRelease(release *types.Release, releaseRollout *types.ReleaseRollout) error {
	// deploy each wave in order
	for ix, wave := range release.Waves {
		// set status to deploying
		releaseRolloutWave := releaseRollout.Waves[ix]
		releaseRolloutWave.Status = types.RolloutStatusDeploying
		if err := setReleaseValue(release); err != nil {
			return err
		}
		// deploy
		if deployReleaseWave(wave, releaseRollout) {
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

func deployReleaseWave(wave *types.ReleaseWave, releaseRollout *types.ReleaseRollout) bool {
	log.Debugf("Deploying wave with %d targets", len(wave.Targets))
	// deploy targets in parallel
	var wg sync.WaitGroup
	var failed bool
	for _, target := range wave.Targets {
		wg.Add(1)
		go func(target *types.ReleaseTarget) {
			defer wg.Done()
			if err := deployReleaseTarget(target, releaseRollout); err != nil {
				log.WithError(err).Error("failed deploying target")
				failed = true
			}
		}(target)
	}
	wg.Wait()
	log.Debugf("Deployed wave with %d targets", len(wave.Targets))
	return failed
}

func deployReleaseTarget(target *types.ReleaseTarget, releaseRollout *types.ReleaseRollout) error {
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
		return rollout.Run(false)
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
		return jobRun.Run(false)
	}
	return nil
}

func syncConfigMaps(target *types.ReleaseTarget, releaseRollout *types.ReleaseRollout) error {
	configmapNames, err := templates.ConfigMaps(releaseRollout.Env, target.Branch)
	if err != nil {
		return err
	}
	for _, configmapName := range configmapNames {
		configmapTemplate, err := templates.ConfigMap(releaseRollout.Env, target.Branch, configmapName)
		if err != nil {
			return err
		}
		configmap := new(v1.ConfigMap)
		err = configmapTemplate.Parse(configmap)
		if err != nil {
			return err
		}
		existingConfigMap, resp, err := kube.ConfigMaps.Get(releaseRollout.Env, configmapName)
		if err != nil {
			return err
		}
		if existingConfigMap == nil {
			_, resp, err = kube.ConfigMaps.Create(releaseRollout.Env, configmap)
		} else {
			_, resp, err = kube.ConfigMaps.Replace(releaseRollout.Env, configmapName, configmap)
		}
		if err != nil {
			return err
		}
		if resp != nil {
			return fmt.Errorf("Got error response from kubernetes: %s", resp)
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
