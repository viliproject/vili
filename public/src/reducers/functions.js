import { CHANGE_FUNCTION, SET_FUNCTION_FIELD } from "../constants"
import Function from "../models/Function"

import { getInitialState, setDataField, changeObject } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case CHANGE_FUNCTION:
      return changeObject(state, action, Function)
    case SET_FUNCTION_FIELD:
      return setDataField(state, action)
    default:
      return state
  }
}
