import { browserHistory } from 'react-router'

import { CHANGE_DEPLOYMENT, SET_DEPLOYMENT_FIELD } from '../constants'

import { subObjects, setDataField } from './utils'

export function subDeployments (env) {
  return subObjects(CHANGE_DEPLOYMENT, 'deployments', env)
}

export function getDeploymentRepository (env, name) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.deployments.getRepository(env, name)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_DEPLOYMENT_FIELD, env, name, 'repository', results.images))
    return { results }
  }
}

export function getDeploymentSpec (env, name) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.deployments.getSpec(env, name)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_DEPLOYMENT_FIELD, env, name, 'spec', results.spec))
    return { results }
  }
}

export function getDeploymentService (env, name) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.deployments.getService(env, name)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_DEPLOYMENT_FIELD, env, name, 'service', results))
    return { results }
  }
}

export function resumeDeployment (env, name) {
  return async function (dispatch, getState, api) {
    return await api.deployments.resume(env, name)
  }
}

export function pauseDeployment (env, name) {
  return async function (dispatch, getState, api) {
    return await api.deployments.pause(env, name)
  }
}

export function scaleDeployment (env, name, replicas) {
  return async function (dispatch, getState, api) {
    return await api.deployments.scale(env, name, replicas)
  }
}

export function deployTag (env, name, tag, branch) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.rollouts.create(env, name, {
      tag: tag,
      branch: branch
    })
    if (error) {
      return { error }
    }
    browserHistory.push(`/${env}/deployments/${name}/rollouts`)
    return { results }
  }
}

export function rollbackToRevision (env, name, toRevision) {
  return async function (dispatch, getState, api) {
    return await api.deployments.rollback(env, name, toRevision)
  }
}
