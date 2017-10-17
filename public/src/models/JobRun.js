import Immutable from 'immutable'

import displayTime from '../lib/displayTime'
import { defaultFields } from './utils'

export default class JobRun extends Immutable.Record({
  ...defaultFields
}) {
  getLabel (key) {
    return this.getIn(['metadata', 'labels', key])
  }

  hasLabel (key, value) {
    return this.getLabel(key) === value
  }

  getAnnotation (key) {
    return this.getIn(['metadata', 'annotations', key])
  }

  get imageTag () {
    return this.getIn(['spec', 'template', 'spec', 'containers', 0, 'image'], ':').split(':')[1]
  }

  get imageBranch () {
    return this.getAnnotation('vili/branch')
  }

  get startedBy () {
    return this.getAnnotation('vili/startedBy')
  }

  get statusName () {
    var status = 'Running'
    this
      .getIn(['status', 'conditions'], Immutable.List())
      .forEach((condition) => {
        switch (condition.get('type')) {
          case 'Complete':
            status = 'Complete'
            break
          case 'Failed':
            status = 'Failed'
            break
        }
      })
    return status
  }

  get revision () {
    const revision = this.getAnnotation('deployment.kubernetes.io/revision')
    return revision ? parseInt(revision) : null
  }

  get creationTimestamp () {
    return new Date(this.getIn(['metadata', 'creationTimestamp']))
  }

  get runAt () {
    return displayTime(this.creationTimestamp)
  }

  get startedAt () {
    return displayTime(new Date(this.getIn(['status', 'startTime'])))
  }

  get completedAt () {
    return displayTime(new Date(this.getIn(['status', 'completionTime'])))
  }
}
