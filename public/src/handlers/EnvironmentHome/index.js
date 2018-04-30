import React from "react"
import PropTypes from "prop-types"

import EnvironmentHome from "../../containers/EnvironmentHome"

export class EnvironmentHomeHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <EnvironmentHome envName={envName} />
  }
}

EnvironmentHomeHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default EnvironmentHomeHandler
