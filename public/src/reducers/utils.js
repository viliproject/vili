import BaseModel from '../models/BaseModel'

export function getInitialState () {
  return {
    envs: {},
    lookUp,
    lookUpData,
    lookUpObject,
    lookUpObjects,
    lookUpObjectsByFunc
  }
}

export function newEnv () {
  return { keys: {} }
}

function lookUp (env) {
  return this.envs[env] || newEnv()
}

function lookUpObjects (env) {
  const envData = this.lookUp(env).keys
  return Object.keys(envData).reduce((obj, key) => {
    obj[key] = envData[key].object
    return obj
  }, {})
}

function lookUpData (env, name) {
  return this.lookUp(env).keys[name] || {}
}

function lookUpObject (env, name) {
  return this.lookUpData(env, name).object
}

function lookUpObjectsByFunc (env, filterFunc) {
  const objects = this.lookUp(env).keys
  return Object.keys(objects).filter(
    (key) => filterFunc(objects[key].object)
  ).reduce((obj, key) => {
    obj[key] = objects[key].object
    return obj
  }, {})
}

export function setEnvField (state, action) {
  const { env, field, data } = action.payload
  const newState = Object.assign({}, state)
  if (!newState.envs[env]) {
    newState.envs[env] = newEnv()
  }
  newState.envs[env][field] = data
  return newState
}

export function setData (state, env, name, data) {
  const newState = Object.assign({}, state)
  if (!newState.envs[env]) {
    newState.envs[env] = newEnv()
  }
  const newStateKeys = newState.envs[env].keys
  if (!newStateKeys[name]) {
    newStateKeys[name] = {}
  }
  newStateKeys[name] = Object.assign({}, newStateKeys[name], data)
  return newState
}

export function setDataField (state, action) {
  const { env, name, field, data } = action.payload
  return setData(state, env, name, {[field]: data})
}

export function deleteData (state, action) {
  const { env, name } = action.payload
  const data = state.lookUpData(env, name)
  if (!data) {
    return state
  }
  const newState = Object.assign({}, state)
  delete newState.envs[env].keys[name]
  return newState
}

function initObjects (state, env, list, Model) {
  const newState = Object.assign({}, state)
  if (!newState.envs[env]) {
    newState.envs[env] = newEnv()
  }
  const newStateKeys = newState.envs[env].keys
  list.items.forEach((item) => {
    const name = item.name || item.metadata.name
    newStateKeys[name] = Object.assign({}, newStateKeys[name], { object: getModelObject(Model, item) })
  })
  return newState
}

export function changeObject (state, action, Model) {
  const { env, event: {type: eventType, object, list} } = action.payload
  switch (eventType) {
    case 'INIT':
      return initObjects(state, env, list, Model)
    case 'ADDED':
    case 'MODIFIED':
      return setData(state, env, object.name || object.metadata.name, { object: getModelObject(Model, object) })
    case 'DELETED':
      return deleteData(state, { payload: { env, name: object.name || object.metadata.name } })
  }
  return state
}

function getModelObject (Model, object) {
  Model = Model || BaseModel
  return new Model(object)
}
