import Immutable from 'immutable'

import displayTime from '../lib/displayTime'
import { defaultFields } from './utils'

export default class ReplicaSet extends Immutable.Record({
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

  get deployedBy () {
    return this.getAnnotation('vili/deployedBy')
  }

  get revision () {
    const revision = this.getAnnotation('deployment.kubernetes.io/revision')
    return revision ? parseInt(revision) : null
  }

  get creationTimestamp () {
    return new Date(this.getIn(['metadata', 'creationTimestamp']))
  }

  get deployedAt () {
    return displayTime(this.creationTimestamp)
  }
}
