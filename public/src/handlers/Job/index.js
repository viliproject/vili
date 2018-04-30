import React from "react"
import PropTypes from "prop-types"

import Job from "../../containers/jobs/Job"

export class JobHandler extends React.Component {
  render() {
    const { env: envName, job: jobName } = this.props.match.params
    return <Job envName={envName} jobName={jobName} />
  }
}

JobHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default JobHandler
