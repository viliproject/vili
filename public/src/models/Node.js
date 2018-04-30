import humanSize from "human-size"
import Immutable from "immutable"

import displayTime from "../lib/displayTime"
import { defaultFields } from "./utils"

export default class Node extends Immutable.Record({
  ...defaultFields,
}) {
  getLabel(key) {
    return this.getIn(["metadata", "labels", key])
  }

  hasLabel(key, value) {
    return this.getLabel(key) === value
  }

  getAnnotation(key) {
    return this.getIn(["metadata", "annotations", key])
  }

  get memory() {
    const match = /(\d+)Ki/g.exec(this.getIn(["status", "capacity", "memory"]))
    if (match) {
      return humanSize(parseInt(match[1]) * 1024, 1)
    }
    return this.getIn(["status", "capacity", "memory"])
  }

  get createdAt() {
    return displayTime(new Date(this.getIn(["metadata", "creationTimestamp"])))
  }
}
