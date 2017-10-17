import { createStore, combineReducers, applyMiddleware, compose } from 'redux'
import { browserHistory } from 'react-router'
import { routerMiddleware } from 'react-router-redux'
import thunk from 'redux-thunk'

import reducers from '../reducers'
import api from '../lib/viliapi'

export default function (initialState) {
  const rootReducer = combineReducers(reducers)

  const store = createStore(rootReducer, initialState, compose(
    applyMiddleware(
      routerMiddleware(browserHistory),
      thunk.withExtraArgument(api)
    ),
    window.devToolsExtension ? window.devToolsExtension() : (f) => f,
  ))

  return store
}
