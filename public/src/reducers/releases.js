import { CHANGE_RELEASE, SET_RELEASE_SPEC } from "../constants"
import Release from "../models/Release"

import { getInitialState, changeObject, setEnvField } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case CHANGE_RELEASE:
      return changeObject(state, action, Release)
    case SET_RELEASE_SPEC:
      return setEnvField(state, action)
    default:
      return state
  }
}
