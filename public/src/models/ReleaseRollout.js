import Immutable from "immutable"

import displayTime from "../lib/displayTime"

export default class ReleaseRollout extends Immutable.Record({
  id: undefined,
  env: undefined,
  rolloutAt: undefined,
  rolloutBy: undefined,
  status: undefined,
  waves: Immutable.List(),
}) {
  get rolloutAtDate() {
    return new Date(this.get("rolloutAt"))
  }

  get rolloutAtHumanize() {
    return displayTime(this.rolloutAtDate)
  }
}
