import BaseModel from './BaseModel'

export default class DeploymentModel extends BaseModel {

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

  get deployedBy () {
    return this.getAnnotation('vili/deployedBy')
  }

  get revision () {
    const revision = this.getAnnotation('deployment.kubernetes.io/revision')
    return revision ? parseInt(revision) : null
  }

}
