import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import DeploymentsList from "../../handlers/DeploymentsList"
import DeploymentContainer from "../../handlers/DeploymentContainer"

export class Deployments extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={DeploymentsList} />
        <Route path={`${prefix}/:deployment`} component={DeploymentContainer} />
      </Switch>
    )
  }
}

Deployments.propTypes = {
  match: PropTypes.object.isRequired,
}

export default Deployments
