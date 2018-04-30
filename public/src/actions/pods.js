import _ from "underscore"

import { CHANGE_POD, ADD_POD_LOG } from "../constants"

import { subObjects } from "./utils"

const podLogSubscriptions = {}

export function subPods(env) {
  return subObjects(CHANGE_POD, "pods", env)
}

export function deletePod(env, name) {
  return async function(dispatch, getState, api) {
    return await api.pods.del(env, name)
  }
}

export function subPodLog(env, name) {
  return async function(dispatch, getState, api) {
    if (!podLogSubscriptions[env]) {
      podLogSubscriptions[env] = {}
    }
    if (podLogSubscriptions[env][name]) {
      return
    }

    let eventBuffer = []
    const dispatchAddPodLog = _.debounce(() => {
      const currentBuffer = eventBuffer
      eventBuffer = []
      dispatch(addPodLog(env, name, currentBuffer))
    }, 200)
    const ws = api.pods.watchLog(
      event => {
        eventBuffer.push(event)
        dispatchAddPodLog()
      },
      env,
      name
    )
    podLogSubscriptions[env][name] = ws
  }
}

export function unsubPodLog(env, name) {
  if (podLogSubscriptions[env] && podLogSubscriptions[env][name]) {
    const ws = podLogSubscriptions[env][name]
    ws.close()
    delete podLogSubscriptions[env][name]
  }
  return {
    type: "NOOP",
    payload: {
      env: env,
      pod: name,
    },
  }
}

export function addPodLog(env, name, eventBuffer) {
  return {
    type: ADD_POD_LOG,
    payload: {
      env,
      name,
      eventBuffer,
    },
  }
}
