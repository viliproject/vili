import React from 'react'
import ReactDOM from 'react-dom'
import { Provider } from 'react-redux'
import { browserHistory } from 'react-router'
import { syncHistoryWithStore } from 'react-router-redux'

import configureStore from './store'
import router from './router'
import { INIT_ENVS } from './constants'

import './less/app.less'

window.addEventListener('DOMContentLoaded', () => {
  let initialState = {}
  if (window.appConfig) {
    initialState.user = window.appConfig.user
    initialState.defaultEnv = window.appConfig.defaultEnv
  }

  const store = configureStore(initialState)
  const history = syncHistoryWithStore(browserHistory, store)

  if (window.appConfig) {
    store.dispatch({type: INIT_ENVS, payload: {envs: window.appConfig.envs}})
  }

  ReactDOM.render(
    <Provider store={store}>
      {router(history)}
    </Provider>, document.getElementById('app'))
})
