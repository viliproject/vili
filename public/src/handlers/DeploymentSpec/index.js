import React from "react"
import PropTypes from "prop-types"

import DeploymentSpec from "../../containers/deployments/DeploymentSpec"

export class DeploymentSpecHandler extends React.Component {
  render() {
    const { env: envName, deployment: deploymentName } = this.props.match.params
    return <DeploymentSpec envName={envName} deploymentName={deploymentName} />
  }
}

DeploymentSpecHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default DeploymentSpecHandler
