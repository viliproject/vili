import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'

export default class ConfigMapModel extends BaseModel {

  get createdAt () {
    return displayTime(new Date(this.metadata.creationTimestamp))
  }

  get keyCount () {
    return Object.keys(this.data).length
  }

}
