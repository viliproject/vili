import React from "react"
import PropTypes from "prop-types"

import JobsList from "../../containers/jobs/JobsList"

export class JobsListHandler extends React.Component {
  render() {
    const { env: envName } = this.props.match.params
    return <JobsList envName={envName} />
  }
}

JobsListHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default JobsListHandler
