import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button } from "react-bootstrap"

import { rollbackToRevision } from "../../actions/deployments"

const dispatchProps = {
  rollbackToRevision,
}
class DeploymentRolloutHistoryRow extends React.Component {
  rollbackTo = event => {
    const { env, deployment, replicaSet } = this.props
    event.target.setAttribute("disabled", "disabled")
    this.props.rollbackToRevision(env, deployment, replicaSet.revision)
  }

  render() {
    const { replicaSet } = this.props
    return (
      <tr>
        <td>{replicaSet.revision}</td>
        <td>{replicaSet.imageTag}</td>
        <td>{replicaSet.imageBranch}</td>
        <td style={{ textAlign: "right" }}>{replicaSet.deployedAt}</td>
        <td style={{ textAlign: "right" }}>{replicaSet.deployedBy}</td>
        <td style={{ textAlign: "right" }}>
          <Button bsStyle="danger" bsSize="xs" onClick={this.rollbackTo}>
            Rollback To
          </Button>
        </td>
      </tr>
    )
  }
}

DeploymentRolloutHistoryRow.propTypes = {
  rollbackToRevision: PropTypes.func,
  env: PropTypes.string,
  deployment: PropTypes.string,
  replicaSet: PropTypes.object,
}

export default connect(null, dispatchProps)(DeploymentRolloutHistoryRow)
