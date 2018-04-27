import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"

import Table from "../../components/Table"
import { activateJobTab } from "../../actions/app"
import { makeLookUpObjectsByLabel } from "../../selectors"

import JobRunsRow from "./JobRunsRow"

function makeMapStateToProps() {
  const lookUpObjectsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { envName, jobName } = ownProps
    const jobRuns = lookUpObjectsByLabel(state.jobRuns, envName, "job", jobName)
    return {
      jobRuns,
    }
  }
}

const dispatchProps = {
  activateJobTab,
}

export class JobRuns extends React.Component {
  componentDidMount() {
    this.props.activateJobTab("runs")
  }

  render() {
    const { envName, jobName, jobRuns } = this.props
    const columns = [
      { title: "Run", key: "run" },
      { title: "Tag", key: "tag", style: { width: "180px" } },
      { title: "Start Time", key: "startTime", style: { width: "180px" } },
      {
        title: "Completion Time",
        key: "completionTime",
        style: { width: "180px" },
      },
      { title: "Status", key: "status", style: { textAlign: "right" } },
      { title: "Actions", key: "actions", style: { textAlign: "right" } },
    ]

    const rows = []
    jobRuns.sortBy(r => -r.creationTimestamp).forEach(jobRun => {
      rows.push({
        component: (
          <JobRunsRow
            key={jobRun.getIn(["metadata", "name"])}
            envName={envName}
            jobName={jobName}
            jobRun={jobRun}
          />
        ),
      })
    })
    return <Table columns={columns} rows={rows} />
  }
}

JobRuns.propTypes = {
  envName: PropTypes.string,
  jobName: PropTypes.string,
  activateJobTab: PropTypes.func.isRequired,
  jobRuns: PropTypes.object,
}

export default connect(makeMapStateToProps, dispatchProps)(JobRuns)
