import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"

import { deleteJobRun } from "../../actions/jobRuns"

const dispatchProps = {
  deleteJobRun,
}

export class JobRunsRow extends React.Component {
  deleteJobRun = () => {
    const { envName, jobRun } = this.props
    this.props.deleteJobRun(envName, jobRun.getIn(["metadata", "name"]))
  }

  render() {
    const { envName, jobName, jobRun } = this.props
    const jobRunName = jobRun.getIn(["metadata", "name"])
    return (
      <tr>
        <td>
          <Link to={`/${envName}/jobs/${jobName}/runs/${jobRunName}`}>
            {jobRunName}
          </Link>
        </td>
        <td>{jobRun.imageTag}</td>
        <td>{jobRun.startedAt}</td>
        <td>{jobRun.completedAt}</td>
        <td style={{ textAlign: "right" }}>{jobRun.statusName}</td>
        <td style={{ textAlign: "right" }}>
          <button
            type="button"
            className="btn btn-xs btn-danger"
            onClick={this.deleteJobRun}
          >
            Delete
          </button>
        </td>
      </tr>
    )
  }
}

JobRunsRow.propTypes = {
  envName: PropTypes.string,
  jobName: PropTypes.string,
  jobRun: PropTypes.object,
  deleteJobRun: PropTypes.func,
}

export default connect(null, dispatchProps)(JobRunsRow)
