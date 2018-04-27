import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import { Panel, Button, Badge } from "react-bootstrap"

import Table from "../../components/Table"
import { deletePod } from "../../actions/pods"

const dispatchProps = {
  deletePod,
}

class DeploymentRolloutPanel extends React.Component {
  deletePod = (event, pod) => {
    event.target.setAttribute("disabled", "disabled")
    this.props.deletePod(this.props.env, pod)
  }

  renderReplicaSetMetadata() {
    const { replicaSet } = this.props
    const metadata = []
    metadata.push(<dt key="t-tag">Tag</dt>)
    metadata.push(<dd key="d-tag">{replicaSet.imageTag}</dd>)
    if (replicaSet.imageBranch) {
      metadata.push(<dt key="t-branch">Branch</dt>)
      metadata.push(<dd key="d-branch">{replicaSet.imageBranch}</dd>)
    }
    metadata.push(<dt key="t-time">Time</dt>)
    metadata.push(<dd key="d-time">{replicaSet.deployedAt}</dd>)
    if (replicaSet.deployedBy) {
      metadata.push(<dt key="t-deployedBy">Deployed By</dt>)
      metadata.push(<dd key="d-deployedBy">{replicaSet.deployedBy}</dd>)
    }
    return (
      <div>
        <dl>{metadata}</dl>
      </div>
    )
  }

  renderPodsTable() {
    const self = this
    const { env, pods } = this.props
    if (pods.isEmpty()) {
      return null
    }
    const columns = [
      { title: "Pod", key: "name" },
      { title: "Host", key: "host" },
      { title: "Phase", key: "phase" },
      { title: "Ready", key: "ready" },
      { title: "Pod IP", key: "podIP" },
      { title: "Created", key: "created" },
      { title: "Actions", key: "actions", style: { textAlign: "right" } },
    ]

    const rows = []
    pods.forEach(pod => {
      const nameLink = (
        <Link to={`/${env}/pods/${pod.getIn(["metadata", "name"])}`}>
          {pod.getIn(["metadata", "name"])}
        </Link>
      )
      const hostLink = (
        <Link to={`/${env}/nodes/${pod.getIn(["spec", "nodeName"])}`}>
          {pod.getIn(["spec", "nodeName"])}
        </Link>
      )
      var actions = (
        <Button
          bsStyle="danger"
          bsSize="xs"
          onClick={event =>
            self.deletePod(event, pod.getIn(["metadata", "name"]))
          }
        >
          Delete
        </Button>
      )
      rows.push({
        key: pod.getIn(["metadata", "name"]),
        name: nameLink,
        host: hostLink,
        phase: pod.getIn(["status", "phase"]),
        ready: pod.isReady ? String.fromCharCode("10003") : "",
        podIP: pod.getIn(["status", "podIP"]),
        created: pod.createdAt,
        actions,
      })
    })

    return <Table columns={columns} rows={rows} fill />
  }

  render() {
    const { deployment, replicaSet, pods } = this.props
    var bsStyle = ""
    if (replicaSet.revision === deployment.revision) {
      bsStyle = "success"
    } else {
      bsStyle = "warning"
    }

    const header = (
      <div>
        <strong>{replicaSet.revision}</strong>
        <Badge bsStyle={bsStyle} pullRight>
          {pods.size}/{replicaSet.getIn(["status", "replicas"])}
        </Badge>
      </div>
    )
    return (
      <Panel key={replicaSet.getIn(["metadata", "name"])} bsStyle={bsStyle}>
        <Panel.Heading>{header}</Panel.Heading>
        <Panel.Body>{this.renderReplicaSetMetadata()}</Panel.Body>
        {this.renderPodsTable()}
      </Panel>
    )
  }
}

DeploymentRolloutPanel.propTypes = {
  deletePod: PropTypes.func.isRequired,
  env: PropTypes.string,
  deployment: PropTypes.object,
  replicaSet: PropTypes.object,
  pods: PropTypes.object,
}

export default connect(null, dispatchProps)(DeploymentRolloutPanel)
