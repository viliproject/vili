import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import JobContainer from "../../containers/jobs/JobContainer"
import Job from "../../handlers/Job"
import JobRuns from "../../handlers/JobRuns"
import JobRun from "../../handlers/JobRun"
import JobSpec from "../../handlers/JobSpec"
import NotFoundPage from "../../components/NotFoundPage"

export class JobContainerHandler extends React.Component {
  render() {
    const prefix = this.props.match.path
    const { env: envName, job: jobName } = this.props.match.params
    return (
      <JobContainer envName={envName} jobName={jobName}>
        <Switch>
          <Route exact path={`${prefix}`} component={Job} />
          <Route exact path={`${prefix}/runs`} component={JobRuns} />
          <Route exact path={`${prefix}/runs/:run`} component={JobRun} />
          <Route exact path={`${prefix}/spec`} component={JobSpec} />
          <Route component={NotFoundPage} />
        </Switch>
      </JobContainer>
    )
  }
}

JobContainerHandler.propTypes = {
  match: PropTypes.object.isRequired,
}

export default JobContainerHandler
