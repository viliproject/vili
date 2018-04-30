import React from "react"
import PropTypes from "prop-types"

import NodesList from "../../containers/nodes/NodesList"

export class NodesListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <NodesList envName={envName} />
  }
}

NodesListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default NodesListHandler
