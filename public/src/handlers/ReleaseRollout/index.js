import React from "react"
import PropTypes from "prop-types"

import ReleaseRollout from "../../containers/releases/ReleaseRollout"

export class ReleaseRolloutHandler extends React.Component {
  render() {
    const {
      env: envName,
      release: releaseName,
      rollout: rolloutID,
    } = this.props.match.params
    return (
      <ReleaseRollout
        envName={envName}
        releaseName={releaseName}
        rolloutID={parseInt(rolloutID)}
      />
    )
  }
}

ReleaseRolloutHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ReleaseRolloutHandler
