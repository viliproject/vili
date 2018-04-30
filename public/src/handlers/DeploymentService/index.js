import React from "react"
import PropTypes from "prop-types"

import DeploymentService from "../../containers/deployments/DeploymentService"

export class DeploymentServiceHandler extends React.Component {
  render() {
    const { env: envName, deployment: deploymentName } = this.props.match.params
    return (
      <DeploymentService envName={envName} deploymentName={deploymentName} />
    )
  }
}

DeploymentServiceHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default DeploymentServiceHandler
