import React from "react"
import PropTypes from "prop-types"

import DeploymentsList from "../../containers/deployments/DeploymentsList"

export class DeploymentsListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <DeploymentsList envName={envName} />
  }
}

DeploymentsListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default DeploymentsListHandler
