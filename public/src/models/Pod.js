import Immutable from "immutable"

import displayTime from "../lib/displayTime"
import { defaultFields } from "./utils"

export default class Pod extends Immutable.Record({
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

  get createdAt() {
    return displayTime(new Date(this.getIn(["metadata", "creationTimestamp"])))
  }

  get imageTag() {
    return this.getIn(["spec", "containers", 0, "image"], ":").split(":")[1]
  }

  get imageBranch() {
    return this.getAnnotation("vili/branch")
  }

  get deployedBy() {
    return this.getAnnotation("vili/deployedBy")
  }

  get isReady() {
    return (
      this.getIn(["status", "phase"]) === "Running" &&
      this.getIn(["status", "containerStatuses"], Immutable.List()).every(cs =>
        cs.get("ready")
      )
    )
  }
}
