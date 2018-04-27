import React from "react"
import PropTypes from "prop-types"

import Release from "../../containers/releases/Release"

export class ReleaseHandler extends React.Component {
  render() {
    const { env: envName, release: releaseName } = this.props.match.params
    return <Release envName={envName} releaseName={releaseName} />
  }
}

ReleaseHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ReleaseHandler
