import Immutable from "immutable"

import displayTime from "../lib/displayTime"
import { defaultFields } from "./utils"

export default class Job extends Immutable.Record({
  ...defaultFields,
}) {
  getLabel(key) {
    return this.getIn(["metadata", "labels", key])
  }

  hasLabel(key, value) {
    return this.getLabel(key) === value
  }

  get creationTimestamp() {
    return new Date(this.getIn(["metadata", "creationTimestamp"]))
  }

  get createdAt() {
    return displayTime(this.creationTimestamp)
  }
}
