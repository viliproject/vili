import {
  SET_ENV,
  ACTIVATE_NAV,
  ACTIVATE_DEPLOYMENT_TAB,
  ACTIVATE_FUNCTION_TAB,
  ACTIVATE_JOB_TAB,
} from "../constants"

export function setEnv(env) {
  return {
    type: SET_ENV,
    payload: {
      env,
    },
  }
}

export function activateNav(item, subItem) {
  return {
    type: ACTIVATE_NAV,
    payload: {
      item,
      subItem,
    },
  }
}

export function activateDeploymentTab(tab) {
  return {
    type: ACTIVATE_DEPLOYMENT_TAB,
    payload: {
      tab,
    },
  }
}

export function activateJobTab(tab) {
  return {
    type: ACTIVATE_JOB_TAB,
    payload: {
      tab,
    },
  }
}

export function activateFunctionTab(tab) {
  return {
    type: ACTIVATE_FUNCTION_TAB,
    payload: {
      tab,
    },
  }
}
