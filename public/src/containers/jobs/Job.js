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

function mapStateToProps (state, ownProps) {
  const { env, job: jobName } = ownProps.params
  const job = state.jobs.lookUpData(env, jobName)
  const jobRuns = state.jobRuns.lookUpObjectsByFunc(env, (obj) => {
    return obj.hasLabel('job', jobName)
  })
  return {
    job,
    jobRuns
  }
}

@connect(mapStateToProps)
export default class Job extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    job: PropTypes.object,
    jobRuns: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateJobTab('home'))
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params } = this.props
    this.props.dispatch(getJobRepository(params.env, params.job))
  }

  render () {
    const { params, job, jobRuns } = this.props
    if (!job) {
      return (<Loading />)
    }

    const tagJobRuns = {}
    _.each(jobRuns, function (jobRun) {
      const tag = jobRun.imageTag
      if (!tagJobRuns[tag]) {
        tagJobRuns[tag] = []
      }
      tagJobRuns[tag].push(jobRun)
    })

    const columns = [
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Branch', key: 'branch', style: {width: '120px'}},
      {title: 'Revision', key: 'revision', style: {width: '90px'}},
      {title: 'Build Time', key: 'buildTime', style: {width: '180px'}},
      {title: 'Run Times', key: 'runtimes', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    var rows = _.map(job.repository, (image) => {
      const buildTime = new Date(image.lastModified)
      const runs = _.sortBy(tagJobRuns[image.tag] || [], x => -x.creationTimestamp)
      return {
        component: (
          <JobRow key={image.tag}
            env={params.env}
            job={params.job}
            tag={image.tag}
            branch={image.branch}
            revision={image.revision}
            buildTime={displayTime(buildTime)}
            jobRuns={runs}
          />),
        time: buildTime.getTime()
      }
    })

    rows = _.sortBy(rows, function (row) {
      return -row.time
    })

    return (<Table columns={columns} rows={rows} />)
  }

}
