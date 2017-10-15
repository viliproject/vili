import { CHANGE_POD, ADD_POD_LOG } from '../constants'
import Pod from '../models/Pod'

import { getInitialState, changeObject, EnvRecord } from './utils'

const initialState = getInitialState()

function addPodLog (state, action) {
  const { env, name, eventBuffer } = action.payload
  return state
    .updateIn(['envs', env], new EnvRecord(), (e) => e)
    .updateIn(['envs', env, 'keys', name, 'log'], '', (l) => {
      eventBuffer.forEach((event) => {
        if (event.type === 'START') {
          l = event.object
        } else {
          l = event.object + '\n' + l
        }
      })
      return l
    })
}

export default function (state = initialState, action) {
  switch (action.type) {
    case CHANGE_POD:
      return changeObject(state, action, Pod)
    case ADD_POD_LOG:
      return addPodLog(state, action)
    default:
      return state
  }
}
