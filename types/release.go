package types

import "time"

// Release is an object representing a release
type Release struct {
	TargetEnv string            `json:"targetEnv"`
	Name      string            `json:"name"`
	Link      string            `json:"link,omitempty"`
	Waves     []*ReleaseWave    `json:"waves"`
	CreatedAt time.Time         `json:"createdAt"`
	CreatedBy string            `json:"createdBy"`
	Rollouts  []*ReleaseRollout `json:"rollouts"`
}

// ReleaseWave represents a wave of a release
type ReleaseWave struct {
	Targets []*ReleaseTarget `json:"targets"`
}

// ReleaseTarget represents a wave of a release
type ReleaseTarget struct {
	Type   ReleaseTargetType `json:"type"`
	Name   string            `json:"name"`
	Branch string            `json:"branch,omitempty"`
	Tag    string            `json:"tag,omitempty"`
}

// ReleaseRollout represents a rollout of a release
type ReleaseRollout struct {
	ID        int                   `json:"id"`
	Env       string                `json:"env"`
	RolloutAt time.Time             `json:"rolloutAt"`
	RolloutBy string                `json:"rolloutBy"`
	Status    RolloutStatus         `json:"status"`
	Waves     []*ReleaseRolloutWave `json:"waves"`
}

// ReleaseRolloutWave represents a wave of a release rollout
type ReleaseRolloutWave struct {
	Status RolloutStatus `json:"status"`
}

// RolloutStatus is the status of the rollout
// It can be one of "new", "deploying", "deployed", "failed"
type RolloutStatus string

// RolloutStatus enum values
const (
	RolloutStatusNew       RolloutStatus = "new"
	RolloutStatusDeploying RolloutStatus = "deploying"
	RolloutStatusDeployed  RolloutStatus = "deployed"
	RolloutStatusFailed    RolloutStatus = "failed"
)

// ReleaseTargetType is the type of the release target
// It can be one of "action", "job" or "app"
type ReleaseTargetType string

// ReleaseTargetType enum values
const (
	ReleaseTargetTypeAction ReleaseTargetType = "action"
	ReleaseTargetTypeJob    ReleaseTargetType = "job"
	ReleaseTargetTypeApp    ReleaseTargetType = "app"
)
