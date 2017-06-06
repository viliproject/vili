import { CHANGE_JOB_RUN } from '../constants'

import { subObjects } from './utils'

export function subJobRuns (env) {
  return subObjects(CHANGE_JOB_RUN, 'jobRuns', env)
}

export function deleteJobRun (env, name) {
  return async function (dispatch, getState, api) {
    return await api.jobRuns.del(env, name)
  }
}
