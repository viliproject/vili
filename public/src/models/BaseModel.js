import _ from 'underscore'

export default class BaseModel {
  constructor (data) {
    _.each(data, (value, key) => {
      Object.defineProperty(this, key, {
        value: value,
        writable: true,
        enumerable: true,
        configurable: true
      })
    })
  }

  updateData = (data) => {
    Object.assign(this, data)
  }

  hasLabel = (key, value) => {
    const label = this.metadata.labels[key]
    return label && label === value
  }
}
