import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button, Label } from "react-bootstrap"

import { deployTag } from "../../actions/functions"

const dispatchProps = {
  deployTag,
}

export class FunctionRow extends React.Component {
  deployTag = event => {
    event.target.setAttribute("disabled", "disabled")
    const { deployTag, env, functionName, tag, branch } = this.props
    deployTag(env, functionName, tag, branch)
  }

  render() {
    const {
      tag,
      branch,
      revision,
      buildTime,
      versions,
      activeVersion,
    } = this.props
    var className = ""
    const deployedAt = []
    versions.forEach(version => {
      var bsStyle = "default"
      if (activeVersion && activeVersion.version === version.version) {
        className = "success"
        bsStyle = "success"
      }
      deployedAt.push(
        <div key={version.version}>
          <Label bsStyle={bsStyle}>
            {version.version} - {version.deployedAt}
          </Label>
        </div>
      )
    })
    return (
      <tr className={className}>
        <td>{tag}</td>
        <td>{branch}</td>
        <td>{revision || "unknown"}</td>
        <td>{buildTime}</td>
        <td style={{ textAlign: "right" }}>{deployedAt}</td>
        <td style={{ textAlign: "right" }}>
          <Button onClick={this.deployTag} bsStyle="primary" bsSize="xs">
            Deploy
          </Button>
        </td>
      </tr>
    )
  }
}

FunctionRow.propTypes = {
  deployTag: PropTypes.func.isRequired,
  env: PropTypes.string,
  functionName: PropTypes.string,
  tag: PropTypes.string,
  branch: PropTypes.string,
  revision: PropTypes.string,
  buildTime: PropTypes.string,
  versions: PropTypes.object,
  activeVersion: PropTypes.object,
}

export default connect(null, dispatchProps)(FunctionRow)
