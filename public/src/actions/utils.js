
export function actionCreator (type, payload) {
  return {
    type,
    payload
  }
}

export function setEnvField (setActionType, env, field, data) {
  return actionCreator(setActionType, { env, field, data })
}

export function setDataField (setActionType, env, name, field, data) {
  return actionCreator(setActionType, { env, name, field, data })
}

export function changeObject (changeActionType, env, event) {
  return actionCreator(changeActionType, { env, event })
}

const objectsSubscriptions = {}

export function subObjects (changeActionType, objectType, env, query) {
  return async function (dispatch, getState, api) {
    if (!objectsSubscriptions[objectType]) {
      objectsSubscriptions[objectType] = {}
    }
    if (objectsSubscriptions[objectType][env]) {
      return
    }
    const ws = api[objectType].watch((data) => {
      dispatch(changeObject(changeActionType, env, data))
    }, env, query)
    objectsSubscriptions[objectType][env] = ws
  }
}
