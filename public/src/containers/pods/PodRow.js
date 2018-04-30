import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import { Button } from "react-bootstrap"

import { deletePod } from "../../actions/pods"

const dispatchProps = {
  deletePod,
}

export class PodRow extends React.Component {
  get nameLink() {
    const { envName, pod } = this.props
    return (
      <Link to={`/${envName}/pods/${pod.getIn(["metadata", "name"])}`}>
        {pod.getIn(["metadata", "name"])}
      </Link>
    )
  }

  get deploymentJobLink() {
    const { envName, pod } = this.props
    if (pod.getLabel("app")) {
      return (
        <Link to={`/${envName}/deployments/${pod.getLabel("app")}`}>
          {pod.getLabel("app")}
        </Link>
      )
    } else if (pod.getLabel("job")) {
      return (
        <Link to={`/${envName}/jobs/${pod.getLabel("job")}`}>
          {pod.getLabel("job")}
        </Link>
      )
    }
  }

  get nodeLink() {
    const { envName, pod } = this.props
    return (
      <Link to={`/${envName}/nodes/${pod.getIn(["spec", "nodeName"])}`}>
        {pod.getIn(["spec", "nodeName"])}
      </Link>
    )
  }

  deletePod = () => {
    const { envName, pod, deletePod } = this.props
    deletePod(envName, pod.getIn(["metadata", "name"]))
  }

  render() {
    const { pod } = this.props
    return (
      <tr>
        <td>{this.nameLink}</td>
        <td>{this.deploymentJobLink}</td>
        <td>{this.nodeLink}</td>
        <td>{pod.getIn(["status", "phase"])}</td>
        <td>{pod.isReady ? String.fromCharCode("10003") : ""}</td>
        <td>{pod.createdAt}</td>
        <td style={{ textAlign: "right" }}>
          <Button onClick={this.deletePod} bsStyle="danger" bsSize="xs">
            Delete
          </Button>
        </td>
      </tr>
    )
  }
}

PodRow.propTypes = {
  deletePod: PropTypes.func.isRequired,
  envName: PropTypes.string,
  pod: PropTypes.object,
}

export default connect(null, dispatchProps)(PodRow)
