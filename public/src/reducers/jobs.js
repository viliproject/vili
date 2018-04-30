import { SET_JOB_FIELD } from "../constants"

import { getInitialState, setDataField } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case SET_JOB_FIELD:
      return setDataField(state, action)
    default:
      return state
  }
}
