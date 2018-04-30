import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button, Label } from "react-bootstrap"
import Immutable from "immutable"

import { runTag } from "../../actions/jobs"

const dispatchProps = {
  runTag,
}

export class JobRow extends React.Component {
  static propTypes = {
    runTag: PropTypes.func.isRequired,
    env: PropTypes.string,
    job: PropTypes.string,
    tag: PropTypes.string,
    branch: PropTypes.string,
    revision: PropTypes.string,
    buildTime: PropTypes.string,
    jobRuns: PropTypes.object,
  }

  runTag = event => {
    event.target.setAttribute("disabled", "disabled")
    const { runTag, env, job, tag, branch } = this.props
    runTag(env, job, tag, branch)
  }

  render() {
    const { tag, branch, revision, buildTime, jobRuns } = this.props
    const runTimes = []
    jobRuns.forEach(jobRun => {
      var bsStyle = "default"
      jobRun
        .getIn(["status", "conditions"], Immutable.List())
        .forEach(condition => {
          switch (condition.get("type")) {
            case "Complete":
              bsStyle = "success"
              break
            case "Failed":
              bsStyle = "danger"
              break
          }
        })
      runTimes.push(
        <div key={jobRun.getIn(["metadata", "name"])}>
          <Label bsStyle={bsStyle}>{jobRun.runAt}</Label>
        </div>
      )
    })
    return (
      <tr>
        <td>{tag}</td>
        <td>{branch}</td>
        <td>{revision || "unknown"}</td>
        <td>{buildTime}</td>
        <td style={{ textAlign: "right" }}>{runTimes}</td>
        <td style={{ textAlign: "right" }}>
          <Button onClick={this.runTag} bsStyle="primary" bsSize="xs">
            Run
          </Button>
        </td>
      </tr>
    )
  }
}

export default connect(null, dispatchProps)(JobRow)
