import React from "react"
import PropTypes from "prop-types"

import JobSpec from "../../containers/jobs/JobSpec"

export class JobSpecHandler extends React.Component {
  render() {
    const { env: envName, job: jobName } = this.props.match.params
    return <JobSpec envName={envName} jobName={jobName} />
  }
}

JobSpecHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default JobSpecHandler
