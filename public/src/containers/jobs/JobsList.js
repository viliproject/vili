import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button, ButtonToolbar } from "react-bootstrap"
import { Link } from "react-router-dom"

import Table from "../../components/Table"
import { activateNav } from "../../actions/app"
import { makeLookUpObjects } from "../../selectors"

function makeMapStateToProps() {
  const lookUpObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { envName } = ownProps
    const env = state.envs.getIn(["envs", envName])
    const jobRuns = lookUpObjects(state.jobRuns, envName)
    return {
      env,
      jobRuns,
    }
  }
}

const dispatchProps = {
  activateNav,
}

export class JobsList extends React.Component {
  componentDidMount() {
    this.props.activateNav("jobs")
  }

  render() {
    const { envName, env, jobRuns } = this.props

    const header = (
      <div className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li className="active">Jobs</li>
        </ol>
      </div>
    )

    if (env.approval) {
      header.push(
        <ButtonToolbar key="toolbar" pullRight>
          <Button onClick={this.release} bsStyle="success" bsSize="small">
            Release
          </Button>
        </ButtonToolbar>
      )
    }

    const columns = [
      { title: "Name", key: "name" },
      { title: "Tag", key: "tag", style: { width: "180px" } },
      {
        title: "Last Run",
        key: "lastRun",
        style: { width: "200px", textAlign: "right" },
      },
    ]

    const rows = []
    env.jobs.forEach(jobName => {
      const jobRun = jobRuns
        .filter(r => r.hasLabel("job", jobName))
        .sortBy(r => -r.creationTimestamp)
        .first()
      rows.push({
        name: <Link to={`/${env.name}/jobs/${jobName}`}>{jobName}</Link>,
        tag: jobRun && jobRun.imageTag,
        lastRun: jobRun && jobRun.runAt,
      })
    })

    return (
      <div>
        {header}
        <Table columns={columns} rows={rows} />
      </div>
    )
  }
}

JobsList.propTypes = {
  envName: PropTypes.string.isRequired,
  env: PropTypes.object,
  jobRuns: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
}

export default connect(makeMapStateToProps, dispatchProps)(JobsList)
