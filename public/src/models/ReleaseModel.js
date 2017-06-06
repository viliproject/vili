import _ from 'underscore'

import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'
import ReleaseRolloutModel from './ReleaseRolloutModel'

export default class ReleaseModel extends BaseModel {

  get createdAtHumanize () {
    return displayTime(new Date(this.createdAt))
  }

  envRollouts (env) {
    if (!this.rollouts) {
      return []
    }
    const rollouts = _.map(
      _.filter(this.rollouts, (rollout) => rollout.env === env),
      (rollout) => new ReleaseRolloutModel(rollout)
    )
    return _.sortBy(rollouts, (rollout) => -rollout.rolloutAtDate)
  }

}
