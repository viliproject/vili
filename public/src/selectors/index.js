import { createSelector } from "reselect"

const getState = state => state
const getEnv = (state, env) => env
const getLabelKey = (state, env, key) => key
const getLabelValue = (state, env, key, value) => value
const getNodeName = (state, env, name) => name

export function makeLookUpObjects() {
  return createSelector([getState, getEnv], (state, env) => {
    return state.lookUpObjects(env)
  })
}

export function makeLookUpObjectsByLabel() {
  return createSelector(
    [getState, getEnv, getLabelKey, getLabelValue],
    (state, env, key, value) => {
      return state.lookUpObjects(env).filter(o => o.hasLabel(key, value))
    }
  )
}

export function makeLookUpObjectsByNodeName() {
  return createSelector(
    [getState, getEnv, getNodeName],
    (state, env, nodeName) => {
      return state
        .lookUpObjects(env)
        .filter(o => o.getIn(["spec", "nodeName"]) === nodeName)
    }
  )
}
