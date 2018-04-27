import { CHANGE_NODE } from "../constants"

import { subObjects } from "./utils"

export function subNodes(env) {
  return subObjects(CHANGE_NODE, "nodes", env)
}

export function setNodeSchedulable(env, name, status) {
  return async function(dispatch, getState, api) {
    return await api.nodes.setSchedulable(env, name, status)
  }
}
