import { CHANGE_CONFIGMAP, SET_CONFIGMAP_FIELD } from "../constants"
import ConfigMap from "../models/ConfigMap"

import { getInitialState, changeObject, setDataField } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case CHANGE_CONFIGMAP:
      return changeObject(state, action, ConfigMap)
    case SET_CONFIGMAP_FIELD:
      return setDataField(state, action)
    default:
      return state
  }
}
