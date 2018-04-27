import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import NodesList from "../../handlers/NodesList"
import Node from "../../handlers/Node"

export class Nodes extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={NodesList} />
        <Route exact path={`${prefix}/:node`} component={Node} />
      </Switch>
    )
  }
}

Nodes.propTypes = {
  match: PropTypes.object.isRequired,
}

export default Nodes
