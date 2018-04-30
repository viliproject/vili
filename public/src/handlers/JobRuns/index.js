import React from "react"
import PropTypes from "prop-types"

import JobRuns from "../../containers/jobs/JobRuns"

export class JobRunsHandler extends React.Component {
  render() {
    const { env: envName, job: jobName } = this.props.match.params
    return <JobRuns envName={envName} jobName={jobName} />
  }
}

JobRunsHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default JobRunsHandler
