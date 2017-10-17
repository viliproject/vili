import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import _ from 'underscore'

import displayTime from '../../lib/displayTime'
import Loading from '../../components/Loading'
import Table from '../../components/Table'
import JobRow from '../../components/jobs/JobRow'
import { activateJobTab } from '../../actions/app'
import { getJobRepository } from '../../actions/jobs'
import { makeLookUpObjectsByLabel } from '../../selectors'

function makeMapStateToProps () {
  const lookUpObjectsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { env, job: jobName } = ownProps.params
    const job = state.jobs.lookUpData(env, jobName)
    const jobRuns = lookUpObjectsByLabel(state.jobRuns, env, 'job', jobName)
    return {
      job,
      jobRuns
    }
  }
}

const dispatchProps = {
  activateJobTab,
  getJobRepository
}

export class Job extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    job: PropTypes.object,
    jobRuns: PropTypes.object,
    activateJobTab: PropTypes.func.isRequired,
    getJobRepository: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateJobTab('home')
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params, getJobRepository } = this.props
    getJobRepository(params.env, params.job)
  }

  render () {
    const { params, job, jobRuns } = this.props
    if (!job || !job.get('repository')) {
      return (<Loading />)
    }

    const columns = [
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Branch', key: 'branch', style: {width: '120px'}},
      {title: 'Revision', key: 'revision', style: {width: '90px'}},
      {title: 'Build Time', key: 'buildTime', style: {width: '180px'}},
      {title: 'Run Times', key: 'runtimes', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    let rows = []
    job.get('repository').forEach((image) => {
      const buildTime = new Date(image.get('lastModified'))
      const runs = jobRuns
        .filter(r => r.imageTag === image.get('tag'))
        .sortBy(r => -r.creationTimestamp)
      rows.push({
        component: (
          <JobRow key={image.get('tag')}
            env={params.env}
            job={params.job}
            tag={image.get('tag')}
            branch={image.get('branch')}
            revision={image.get('revision')}
            buildTime={displayTime(buildTime)}
            jobRuns={runs}
          />),
        time: buildTime.getTime()
      })
    })

    rows = _.sortBy(rows, function (row) {
      return -row.time
    })

    return (<Table columns={columns} rows={rows} />)
  }
}

export default connect(makeMapStateToProps, dispatchProps)(Job)
