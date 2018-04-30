import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import PodsList from "../../handlers/PodsList"
import Pod from "../../handlers/Pod"

export class Pods extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={PodsList} />
        <Route exact path={`${prefix}/:pod`} component={Pod} />
      </Switch>
    )
  }
}

Pods.propTypes = {
  match: PropTypes.object.isRequired,
}

export default Pods
