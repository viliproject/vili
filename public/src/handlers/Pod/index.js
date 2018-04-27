import React from "react"
import PropTypes from "prop-types"

import Pod from "../../containers/pods/Pod"

export class PodHandler extends React.Component {
  render() {
    const { env: envName, pod: podName } = this.props.match.params
    return <Pod envName={envName} podName={podName} />
  }
}

PodHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default PodHandler
