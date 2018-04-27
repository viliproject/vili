import Immutable from "immutable"
import {
  SET_ENV,
  ACTIVATE_NAV,
  ACTIVATE_DEPLOYMENT_TAB,
  ACTIVATE_JOB_TAB,
} from "../constants"

const initialState = Immutable.fromJS({
  deploymentTab: "home",
  jobTab: "home",
})

export default function(state = initialState, action) {
  switch (action.type) {
    case SET_ENV:
      return state.set("env", action.payload.env)
    case ACTIVATE_NAV:
      return state.set("nav", action.payload)
    case ACTIVATE_DEPLOYMENT_TAB:
      return state.set("deploymentTab", action.payload.tab)
    case ACTIVATE_JOB_TAB:
      return state.set("jobTab", action.payload.tab)
    default:
      return state
  }
}
