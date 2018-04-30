import history from "../lib/history"
import { SET_JOB_FIELD } from "../constants"

import { setDataField } from "./utils"

export function getJobRepository(env, name) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.jobs.getRepository(env, name)
    if (error) {
      return { error }
    }
    dispatch(
      setDataField(SET_JOB_FIELD, env, name, "repository", results.images)
    )
    return { results }
  }
}

export function getJobSpec(env, name) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.jobs.getSpec(env, name)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_JOB_FIELD, env, name, "spec", results.spec))
    return { results }
  }
}

export function runTag(env, name, tag, branch) {
  return async function(dispatch, getState, api) {
    const { results, error } = await api.jobRuns.create(env, name, {
      tag: tag,
      branch: branch,
    })
    if (error) {
      return { error }
    }
    history.push(`/${env}/jobs/${name}/runs/${results.job.metadata.name}`)
    return { results }
  }
}
