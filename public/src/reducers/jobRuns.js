import { CHANGE_JOB_RUN } from '../constants'
import JobRun from '../models/JobRun'

import { getInitialState, changeObject } from './utils'

const initialState = getInitialState()

export default function (state = initialState, action) {
  switch (action.type) {
    case CHANGE_JOB_RUN:
      return changeObject(state, action, JobRun)
    default:
      return state
  }
}
