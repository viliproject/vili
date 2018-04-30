import Immutable from "immutable"

export default class Environment extends Immutable.Record({
  name: "",
  branch: "",
  repositoryBranches: Immutable.List(),
  autodeployBranches: Immutable.List(),
  protected: false,
  deployedToEnv: "",
  approvedFromEnv: "",
  jobs: Immutable.List(),
  deployments: Immutable.List(),
  configmaps: Immutable.List(),
}) {}
