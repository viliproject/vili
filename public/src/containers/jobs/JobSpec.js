import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"

import Loading from "../../components/Loading"
import { activateJobTab } from "../../actions/app"
import { getJobSpec } from "../../actions/jobs"

function mapStateToProps(state, ownProps) {
  const { envName, jobName } = ownProps
  const job = state.jobs.lookUpData(envName, jobName)
  return {
    job,
  }
}

const dispatchProps = {
  activateJobTab,
  getJobSpec,
}

export class JobSpec extends React.Component {
  componentDidMount() {
    this.props.activateJobTab("spec")
    this.fetchData()
  }

  componentDidUpdate(prevProps) {
    if (
      this.props.envName !== prevProps.envName ||
      this.props.jobName !== prevProps.jobName
    ) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { envName, jobName } = this.props
    this.props.getJobSpec(envName, jobName)
  }

  render() {
    const { job } = this.props
    if (!job || !job.get("spec")) {
      return <Loading />
    }
    return (
      <div className="col-md-8">
        <div id="source-yaml">
          <pre>
            <code>{job.get("spec")}</code>
          </pre>
        </div>
      </div>
    )
  }
}

JobSpec.propTypes = {
  envName: PropTypes.string.isRequired,
  jobName: PropTypes.string.isRequired,
  job: PropTypes.object,
  activateJobTab: PropTypes.func.isRequired,
  getJobSpec: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(JobSpec)
