import _ from 'underscore'

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
    let initialized = false
    const initialItems = []
    const initialDispatch = _.debounce((event) => {
      initialized = true
      dispatch(changeObject(changeActionType, env, {
        type: 'INIT',
        list: initialItems
      }))
    }, 200)
    const ws = api[objectType].watch((event) => {
      if (!initialized && event.type === 'INIT') {
        initialized = true
      }
      if (initialized) {
        dispatch(changeObject(changeActionType, env, event))
      } else {
        initialItems.push(event.object)
        initialDispatch()
      }
    }, env, query)
    objectsSubscriptions[objectType][env] = ws
  }
}
