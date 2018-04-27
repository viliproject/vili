import React from "react"
import PropTypes from "prop-types"

import JobRun from "../../containers/jobs/JobRun"

export class JobRunHandler extends React.Component {
  render() {
    const { env: envName, job: jobName, run: runName } = this.props.match.params
    return <JobRun envName={envName} jobName={jobName} runName={runName} />
  }
}

JobRunHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default JobRunHandler
