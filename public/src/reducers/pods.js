import { CHANGE_POD, ADD_POD_LOG } from '../constants'
import PodModel from '../models/PodModel'

import { getInitialState, changeObject, newEnv } from './utils'

const initialState = getInitialState()

function addPodLog (state, action) {
  const { env, name, eventBuffer } = action.payload
  const newState = Object.assign({}, state)
  if (!newState.envs[env]) {
    newState.envs[env] = newEnv()
  }
  const newStateKeys = newState.envs[env].keys
  if (!newStateKeys[name]) {
    newStateKeys[name] = {}
  }
  const podData = newStateKeys[name]
  eventBuffer.forEach((event) => {
    if (event.type === 'START') {
      podData.log = event.object
    } else {
      podData.log = event.object + '\n' + podData.log
    }
  })
  return newState
}

export default function (state = initialState, action) {
  switch (action.type) {
    case CHANGE_POD:
      return changeObject(state, action, PodModel)
    case ADD_POD_LOG:
      return addPodLog(state, action)
    default:
      return state
  }
}
