import history from "../lib/history"
import { CHANGE_FUNCTION, SET_FUNCTION_FIELD } from "../constants"

import { subObjects, setDataField } from "./utils"

export function subFunctions(env) {
  return subObjects(CHANGE_FUNCTION, "functions", env)
}

export function getFunctionRepository(env, name) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.functions.getRepository(env, name)
    if (error) {
      return { error }
    }
    dispatch(
      setDataField(SET_FUNCTION_FIELD, env, name, "repository", results.images)
    )
    return { results }
  }
}

export function getFunctionSpec(env, name) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.functions.getSpec(env, name)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_FUNCTION_FIELD, env, name, "spec", results.spec))
    return { results }
  }
}

export function deployTag(env, name, tag, branch) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.functions.deploy(env, name, {
      tag: tag,
      branch: branch,
    })
    if (error) {
      return { error }
    }
    history.push(`/${env}/functions/${name}/versions`)
    return { results }
  }
}

export function rollbackToVersion(env, name, toVersion) {
  return async function(dispatch, getState, api) {
    return await api.functions.rollback(env, name, toVersion)
  }
}
