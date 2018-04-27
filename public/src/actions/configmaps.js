import { CHANGE_CONFIGMAP, SET_CONFIGMAP_FIELD } from "../constants"

import { subObjects, setDataField } from "./utils"

export function subConfigMaps(env) {
  return subObjects(CHANGE_CONFIGMAP, "configmaps", env)
}

export function getConfigMapSpec(env, name) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.configmaps.getSpec(env, name)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_CONFIGMAP_FIELD, env, name, "spec", results))
    return { results }
  }
}

export function createConfigMap(env, name) {
  return async function(dispatch, getState, api) {
    return await api.configmaps.create(env, name)
  }
}

export function deleteConfigMap(env, name) {
  return async function(dispatch, getState, api) {
    return await api.configmaps.del(env, name)
  }
}

export function setConfigMapKeys(env, name, values) {
  return async function(dispatch, getState, api) {
    return await api.configmaps.setKeys(env, name, values)
  }
}

export function deleteConfigMapKey(env, name, key) {
  return async function(dispatch, getState, api) {
    return await api.configmaps.delKey(env, name, key)
  }
}
