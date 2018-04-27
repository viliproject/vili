import React from "react"
import PropTypes from "prop-types"

import DeploymentRollouts from "../../containers/deployments/DeploymentRollouts"

export class DeploymentRolloutsHandler extends React.Component {
  render() {
    const { env: envName, deployment: deploymentName } = this.props.match.params
    return (
      <DeploymentRollouts envName={envName} deploymentName={deploymentName} />
    )
  }
}

DeploymentRolloutsHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default DeploymentRolloutsHandler
