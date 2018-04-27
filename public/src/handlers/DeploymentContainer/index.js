import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import DeploymentContainer from "../../containers/deployments/DeploymentContainer"
import Deployment from "../../handlers/Deployment"
import DeploymentRollouts from "../../handlers/DeploymentRollouts"
import DeploymentSpec from "../../handlers/DeploymentSpec"
import DeploymentService from "../../handlers/DeploymentService"
import NotFoundPage from "../../components/NotFoundPage"

export class DeploymentContainerHandler extends React.Component {
  render() {
    const prefix = this.props.match.path
    const { env: envName, deployment: deploymentName } = this.props.match.params
    return (
      <DeploymentContainer envName={envName} deploymentName={deploymentName}>
        <Switch>
          <Route exact path={`${prefix}`} component={Deployment} />
          <Route
            exact
            path={`${prefix}/rollouts`}
            component={DeploymentRollouts}
          />
          <Route exact path={`${prefix}/spec`} component={DeploymentSpec} />
          <Route
            exact
            path={`${prefix}/service`}
            component={DeploymentService}
          />
          <Route component={NotFoundPage} />
        </Switch>
      </DeploymentContainer>
    )
  }
}

DeploymentContainerHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default DeploymentContainerHandler
