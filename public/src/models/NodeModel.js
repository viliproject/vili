import humanSize from 'human-size'

import displayTime from '../lib/displayTime'

import BaseModel from './BaseModel'

export default class NodeModel extends BaseModel {

  get memory () {
    const match = /(\d+)Ki/g.exec(this.status.capacity.memory)
    if (match) {
      return humanSize(parseInt(match[1]) * 1024, 1)
    }
    return this.status.capacity.memory
  }

  get createdAt () {
    return displayTime(new Date(this.metadata.creationTimestamp))
  }

}
