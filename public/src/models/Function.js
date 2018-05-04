import Immutable from "immutable"

import displayTime from "../lib/displayTime"

export class FunctionVersion extends Immutable.Record({
  tag: "",
  branch: "",
  version: "",
  lastModified: "",
  deployedBy: "",
}) {
  get deployedAt() {
    return displayTime(this.lastModified)
  }
}

export default class Function extends Immutable.Record({
  name: "",
  activeVersion: null,
  versions: Immutable.List(),
}) {
  constructor(obj) {
    super(
      obj
        .update(
          "activeVersion",
          activeVersion =>
            activeVersion && new FunctionVersion().merge(activeVersion)
        )
        .update("versions", versions =>
          versions.map(v => new FunctionVersion().merge(v))
        )
    )
  }
}
