import { ADD_ENV, REMOVE_ENV, SHOW_CREATE_ENV_MODAL, HIDE_CREATE_ENV_MODAL, SET_BRANCHES, SET_ENV_FIELD } from '../constants'

import { actionCreator, setDataField } from './utils'

export function showCreateEnvModal () {
  return {
    type: SHOW_CREATE_ENV_MODAL
  }
}

export function hideCreateEnvModal () {
  return {
    type: HIDE_CREATE_ENV_MODAL
  }
}

export function getBranches () {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.branches.get()
    if (error) {
      return { error }
    }
    dispatch({
      type: SET_BRANCHES,
      payload: results
    })
    return { results }
  }
}

export function getEnvironmentSpec (name, branch) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.environments.getSpec(name, branch)
    if (error) {
      return { error }
    }
    dispatch(setDataField(SET_ENV_FIELD, name, 'spec', results))
    return { results }
  }
}

export function createEnvironment (spec) {
  return async function (dispatch, getState, api) {
    const { results, error } = await api.environments.create(spec)
    if (error) {
      return { error }
    }
    dispatch(actionCreator(ADD_ENV, results.environment))
    return { results }
  }
}

export function deleteEnvironment (name) {
  return async function (dispatch, getState, api) {
    var checkName = prompt('Are you sure you wish to delete this environment? Enter the environment name to confirm')
    if (checkName !== name) {
      return
    }
    const { results, error } = await api.environments.del(name)
    if (error) {
      return { error }
    }
    dispatch(actionCreator(REMOVE_ENV, { name }))
    return { results }
  }
}
