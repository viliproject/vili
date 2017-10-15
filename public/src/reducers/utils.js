import Immutable from 'immutable'

class APIStoreRecord extends Immutable.Record({
  envs: Immutable.Map()
}) {
  lookUp (env) {
    return this.getIn(['envs', env], new EnvRecord())
  }

  lookUpObjects (env) {
    return this.lookUp(env).get('keys').map((v) => v.get('object'))
  }

  lookUpData (env, name) {
    return this.lookUp(env).getIn(['keys', name], Immutable.Map())
  }

  lookUpObject (env, name) {
    return this.lookUp(env).getIn(['keys', name, 'object'])
  }
}

export class EnvRecord extends Immutable.Record({
  keys: Immutable.Map(),
  spec: undefined
}) {
}

export function getInitialState () {
  return new APIStoreRecord()
}

export function setEnvField (state, action) {
  const { env, field, data } = action.payload
  return state.updateIn(['envs', env], new EnvRecord(), (e) => e.set(field, data))
}

export function setData (state, env, name, data) {
  return state.updateIn(['envs', env], new EnvRecord(), (e) => e.mergeIn(['keys', name], data))
}

export function setDataField (state, action) {
  const { env, name, field, data } = action.payload
  return setData(state, env, name, {[field]: data})
}

export function deleteData (state, action) {
  const { env, name } = action.payload
  return state.deleteIn(['envs', env, 'keys', name])
}

function initObjects (state, env, list, Model) {
  if (!state.getIn(['envs', env])) {
    state = state.setIn(['envs', env], new EnvRecord())
  }
  list.forEach((item) => {
    const name = item.name || item.metadata.name
    state = state.setIn(['envs', env, 'keys', name, 'object'], getModelInstance(Model, item))
  })
  return state
}

export function changeObject (state, action, Model) {
  const { env, event: {type: eventType, object, list} } = action.payload
  switch (eventType) {
    case 'INIT':
      return initObjects(state, env, list, Model)
    case 'ADDED':
    case 'MODIFIED':
      return setData(state, env, object.name || object.metadata.name, { object: getModelInstance(Model, object) })
    case 'DELETED':
      return deleteData(state, { payload: { env, name: object.name || object.metadata.name } })
  }
  return state
}

function getModelInstance (Model, object) {
  return new Model(Immutable.fromJS(object))
}
