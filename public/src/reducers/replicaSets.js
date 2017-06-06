import { CHANGE_REPLICA_SET } from '../constants'
import ReplicaSetModel from '../models/ReplicaSetModel'

import { getInitialState, changeObject } from './utils'

const initialState = getInitialState()

export default function (state = initialState, action) {
  switch (action.type) {
    case CHANGE_REPLICA_SET:
      return changeObject(state, action, ReplicaSetModel)
    default:
      return state
  }
}
