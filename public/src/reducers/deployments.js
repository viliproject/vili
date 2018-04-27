import { CHANGE_DEPLOYMENT, SET_DEPLOYMENT_FIELD } from "../constants"
import Deployment from "../models/Deployment"

import { getInitialState, setDataField, changeObject } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case CHANGE_DEPLOYMENT:
      return changeObject(state, action, Deployment)
    case SET_DEPLOYMENT_FIELD:
      return setDataField(state, action)
    default:
      return state
  }
}
