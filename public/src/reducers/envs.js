import Immutable from 'immutable'
import _ from 'underscore'

import { INIT_ENVS, ADD_ENV, REMOVE_ENV, SHOW_CREATE_ENV_MODAL, HIDE_CREATE_ENV_MODAL, SET_BRANCHES } from '../constants'

const initialState = Immutable.fromJS({
  envs: [],
  showCreateModal: false
})

function addEnv (state, payload) {
  const envs = state.get('envs')
  envs.push(payload)
  return state.set('envs', envs)
}

function removeEnv (state, payload) {
  let envs = state.get('envs')
  envs = _.filter(envs, (env) => {
    return env.name !== payload.name
  })
  return state.set('envs', envs)
}

export default function (state = initialState, action) {
  switch (action.type) {
    case INIT_ENVS:
      return state.set('envs', action.payload.envs)
    case ADD_ENV:
      return addEnv(state, action.payload)
    case REMOVE_ENV:
      return removeEnv(state, action.payload)
    case SHOW_CREATE_ENV_MODAL:
      return state.set('showCreateModal', true)
    case HIDE_CREATE_ENV_MODAL:
      return state.set('showCreateModal', false)
    case SET_BRANCHES:
      return state.set('branches', action.payload.branches)
    default:
      return state
  }
}
