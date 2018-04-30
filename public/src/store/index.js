import { createStore, combineReducers, applyMiddleware, compose } from "redux"
import thunk from "redux-thunk"

import reducers from "../reducers"
import api from "../lib/viliapi"

let initialState = {}

if (window.appConfig) {
  initialState.user = window.appConfig.user
  initialState.defaultEnv = window.appConfig.defaultEnv
}

const rootReducer = combineReducers(reducers)

const store = createStore(
  rootReducer,
  initialState,
  compose(
    applyMiddleware(thunk.withExtraArgument(api)),
    window.devToolsExtension ? window.devToolsExtension() : f => f
  )
)

export default store
