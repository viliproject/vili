import Immutable from 'immutable'
import * as Constants from '../constants'

const initialState = Immutable.fromJS({
  deploymentTab: 'home',
  jobTab: 'home'
})

export default function (state = initialState, action) {
  switch (action.type) {
    case Constants.ACTIVATE_NAV:
      return state.set('nav', action.payload)
    case Constants.ACTIVATE_DEPLOYMENT_TAB:
      return state.set('deploymentTab', action.payload.tab)
    case Constants.ACTIVATE_JOB_TAB:
      return state.set('jobTab', action.payload.tab)
    default:
      return state
  }
}
