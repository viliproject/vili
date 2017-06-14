import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'

export default class JobRunModel extends BaseModel {

  getAnnotation (key) {
    if (!this.metadata || !this.metadata.annotations) {
      return null
    }
    return this.metadata.annotations[key]
  }

  getLabel (key) {
    if (!this.metadata || !this.metadata.labels) {
      return null
    }
    return this.metadata.labels[key]
  }

  get imageTag () {
    return this.spec.template.spec.containers[0].image.split(':')[1]
  }

  get imageBranch () {
    return this.getAnnotation('vili/branch')
  }

  get startedBy () {
    return this.getAnnotation('vili/startedBy')
  }

  get statusName () {
    var status = 'Running'
    if (this.status.conditions) {
      this.status.conditions.forEach((condition) => {
        switch (condition.type) {
          case 'Complete':
            status = 'Complete'
            break
          case 'Failed':
            status = 'Failed'
            break
        }
      })
    }
    return status
  }

  get revision () {
    if (!this.metadata || !this.metadata.annotations) {
      return null
    }
    return parseInt(this.metadata.annotations['deployment.kubernetes.io/revision'])
  }

  get creationTimestamp () {
    return new Date(this.metadata.creationTimestamp)
  }

  get runAt () {
    return displayTime(this.creationTimestamp)
  }

  get startedAt () {
    return displayTime(new Date(this.status.startTime))
  }

  get completedAt () {
    return displayTime(new Date(this.status.completionTime))
  }

}
