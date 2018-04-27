import React from "react"
import { connect } from "react-redux"
import { Route, Switch } from "react-router"

import TopNav from "../../containers/TopNav"
import EnvCreateModal from "../../containers/EnvCreateModal"
import SideNav from "../../components/SideNav"

import Home from "../Home"
import Environment from "../Environment"

function mapStateToProps(state) {
  return {
    app: state.app,
  }
}

export class App extends React.Component {
  render() {
    return (
      <div className="top-nav container-fluid full-height">
        <TopNav />
        <div className="page-wrapper">
          <div className="sidebar">
            <SideNav />
          </div>
          <div className="content-wrapper">
            <Switch>
              <Route exact path="/" component={Home} />
              <Route path="/:env" component={Environment} />
            </Switch>
          </div>
        </div>
        <EnvCreateModal />
      </div>
    )
  }
}

App.propTypes = {}

export default connect(mapStateToProps)(App)
