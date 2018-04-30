import React from "react"
import { Router } from "react-router-dom"
import { Route } from "react-router"

import App from "./handlers/App"
import history from "./lib/history"

export default (
  <Router history={history}>
    <Route path="/" component={App} />
  </Router>
)
