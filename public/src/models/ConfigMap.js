import Immutable from 'immutable'

import displayTime from '../lib/displayTime'
import { defaultFields } from './utils'

export default class ConfigMap extends Immutable.Record({
  ...defaultFields,
  data: Immutable.Map()
}) {
  getLabel (key) {
    return this.getIn(['metadata', 'labels', key])
  }

  hasLabel (key, value) {
    return this.getLabel(key) === value
  }

  get createdAt () {
    return displayTime(new Date(this.getIn(['metadata', 'creationTimestamp'])))
  }

  get keyCount () {
    return this.data.size
  }
}
