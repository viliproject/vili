import React from "react"
import PropTypes from "prop-types"

import PodsList from "../../containers/pods/PodsList"

export class PodsListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <PodsList envName={envName} />
  }
}

PodsListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default PodsListHandler
