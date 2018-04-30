import React from "react"
import PropTypes from "prop-types"

import Deployment from "../../containers/deployments/Deployment"

export class DeploymentHandler extends React.Component {
  render() {
    const { env: envName, deployment: deploymentName } = this.props.match.params
    return <Deployment envName={envName} deploymentName={deploymentName} />
  }
}

DeploymentHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default DeploymentHandler
