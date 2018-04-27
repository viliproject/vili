import React from "react"
import ReactDOM from "react-dom"
import { Provider } from "react-redux"

import store from "./store"
import router from "./router"
import { INIT_ENVS } from "./constants"

import "./less/app.less"

window.addEventListener("DOMContentLoaded", () => {
  if (window.appConfig) {
    store.dispatch({
      type: INIT_ENVS,
      payload: { envs: window.appConfig.envs },
    })
  }

  ReactDOM.render(
    <Provider store={store}>{router}</Provider>,
    document.getElementById("app")
  )
})
