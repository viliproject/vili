import BaseModel from './BaseModel'

export default class DeploymentModel extends BaseModel {

  get imageTag () {
    return this.spec.template.spec.containers[0].image.split(':')[1]
  }

  get revision () {
    if (!this.metadata || !this.metadata.annotations) {
      return null
    }
    return parseInt(this.metadata.annotations['deployment.kubernetes.io/revision'])
  }

}
