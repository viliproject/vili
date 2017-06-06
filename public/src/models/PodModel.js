import _ from 'underscore'

import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'

export default class PodModel extends BaseModel {

  get createdAt () {
    return displayTime(new Date(this.metadata.creationTimestamp))
  }

  get imageTag () {
    return this.spec.containers[0].image.split(':')[1]
  }

  get isReady () {
    return this.status.phase === 'Running' &&
           _.every(this.status.containerStatuses, (cs) => cs.ready)
  }

}
