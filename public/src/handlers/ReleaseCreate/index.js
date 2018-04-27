import React from "react"
import PropTypes from "prop-types"

import ReleaseCreate from "../../containers/releases/ReleaseCreate"

export class ReleaseCreateHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <ReleaseCreate envName={envName} />
  }
}

ReleaseCreateHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ReleaseCreateHandler
