import Immutable from "immutable"

import {
  INIT_ENVS,
  ADD_ENV,
  REMOVE_ENV,
  SHOW_CREATE_ENV_MODAL,
  HIDE_CREATE_ENV_MODAL,
  SET_BRANCHES,
} from "../constants"
import Environment from "../models/Environment"

const initialState = Immutable.fromJS({
  envs: Immutable.OrderedMap(),
  showCreateModal: false,
})

function initEnvs(state, payload) {
  payload.envs.forEach(env => {
    state = addEnv(state, env)
  })
  return state
}

function addEnv(state, data) {
  const env = new Environment(Immutable.fromJS(data))
  return state.setIn(["envs", env.name], env)
}

function removeEnv(state, payload) {
  return state.deleteIn(["envs", payload.name])
}

export default function(state = initialState, action) {
  switch (action.type) {
    case INIT_ENVS:
      return initEnvs(state, action.payload)
    case ADD_ENV:
      return addEnv(state, action.payload)
    case REMOVE_ENV:
      return removeEnv(state, action.payload)
    case SHOW_CREATE_ENV_MODAL:
      return state.set("showCreateModal", true)
    case HIDE_CREATE_ENV_MODAL:
      return state.set("showCreateModal", false)
    case SET_BRANCHES:
      return state.set("branches", Immutable.fromJS(action.payload.branches))
    default:
      return state
  }
}
