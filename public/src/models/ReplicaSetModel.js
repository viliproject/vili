import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'

export default class ReplicaSetModel extends BaseModel {

  get imageTag () {
    return this.spec.template.spec.containers[0].image.split(':')[1]
  }

  get imageBranch () {
    return this.metadata.labels['branch']
  }

  get revision () {
    if (!this.metadata || !this.metadata.annotations) {
      return null
    }
    return parseInt(this.metadata.annotations['deployment.kubernetes.io/revision'])
  }

  get deployedAt () {
    return displayTime(new Date(this.metadata.creationTimestamp))
  }

}
