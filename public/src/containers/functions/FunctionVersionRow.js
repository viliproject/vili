import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button } from "react-bootstrap"

import { rollbackToVersion } from "../../actions/functions"

const dispatchProps = {
  rollbackToVersion,
}
class FunctionVersionRow extends React.Component {
  rollbackTo = event => {
    const { env, func, version, rollbackToVersion } = this.props
    event.target.setAttribute("disabled", "disabled")
    rollbackToVersion(env, func, version.version)
  }

  render() {
    const { isActive, version } = this.props
    return (
      <tr className={isActive ? "success" : ""}>
        <td>{version.version}</td>
        <td>{version.tag}</td>
        <td>{version.branch}</td>
        <td style={{ textAlign: "right" }}>{version.deployedAt}</td>
        <td style={{ textAlign: "right" }}>{version.deployedBy}</td>
        <td style={{ textAlign: "right" }}>
          <Button bsStyle="danger" bsSize="xs" onClick={this.rollbackTo}>
            Rollback To
          </Button>
        </td>
      </tr>
    )
  }
}

FunctionVersionRow.propTypes = {
  rollbackToVersion: PropTypes.func,
  env: PropTypes.string,
  func: PropTypes.string,
  version: PropTypes.object,
  isActive: PropTypes.bool,
}

export default connect(null, dispatchProps)(FunctionVersionRow)
