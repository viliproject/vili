import { browserHistory } from 'react-router'

import { CHANGE_RELEASE, SET_RELEASE_SPEC } from '../constants'

import { subObjects, setEnvField } from './utils'

export function subReleases (env) {
  return subObjects(CHANGE_RELEASE, 'releases', env)
}

export function getReleaseSpec (env) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.releases.getSpec(env)
    if (error) {
      return { error }
    }
    dispatch(setEnvField(SET_RELEASE_SPEC, env, 'spec', results))
    return { results }
  }
}

export function createRelease (env, spec) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.releases.create(env, spec)
    if (error) {
      return { error }
    }
    browserHistory.push(`/${env}/releases/${spec.name}`)
    return { results }
  }
}

export function createReleaseFromLatest (env) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.releases.createFromLatest(env)
    if (error) {
      return { error }
    }
    browserHistory.push(`/${env}/releases/${results.name}`)
    return { results }
  }
}

export function deployRelease (env, name) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.releases.deploy(env, name)
    if (error) {
      return { error }
    }
    browserHistory.push(`/${env}/releases/${name}/rollouts/${results.id}`)
    return { results }
  }
}

export function deleteRelease (env, name) {
  return async function (dispatch, getState, api) {
    return await api.releases.del(env, name)
  }
}
