import React from "react"
import PropTypes from "prop-types"

import Node from "../../containers/nodes/Node"

export class NodeHandler extends React.Component {
  render() {
    const { env: envName, node: nodeName } = this.props.match.params
    return <Node envName={envName} nodeName={nodeName} />
  }
}

NodeHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default NodeHandler
