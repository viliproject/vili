import { CHANGE_REPLICA_SET } from "../constants"
import ReplicaSet from "../models/ReplicaSet"

import { getInitialState, changeObject } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case CHANGE_REPLICA_SET:
      return changeObject(state, action, ReplicaSet)
    default:
      return state
  }
}
