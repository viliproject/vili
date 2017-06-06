import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'

export default class ReleaseRolloutModel extends BaseModel {

  get rolloutAtDate () {
    return new Date(this.rolloutAt)
  }

  get rolloutAtHumanize () {
    return displayTime(this.rolloutAtDate)
  }

}
