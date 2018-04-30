import React from "react"
import PropTypes from "prop-types"

import ReleasesList from "../../containers/releases/ReleasesList"

export class ReleasesListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <ReleasesList envName={envName} />
  }
}

ReleasesListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default ReleasesListHandler
