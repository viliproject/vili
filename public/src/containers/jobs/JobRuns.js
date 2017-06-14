import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateJobTab } from '../../actions/app'
import { deleteJobRun } from '../../actions/jobRuns'

function mapStateToProps (state, ownProps) {
  const { env, job } = ownProps.params
  const jobRuns = state.jobRuns.lookUpObjectsByFunc(env, (obj) => {
    return obj.hasLabel('job', job)
  })
  return {
    jobRuns
  }
}

@connect(mapStateToProps)
export default class JobRuns extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    jobRuns: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateJobTab('runs'))
  }

  render () {
    const { params, jobRuns } = this.props
    const sortedRuns = _.sortBy(jobRuns, x => -x.creationTimestamp)
    const columns = [
      {title: 'Run', key: 'run'},
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Start Time', key: 'startTime', style: {width: '180px'}},
      {title: 'Completion Time', key: 'completionTime', style: {width: '180px'}},
      {title: 'Status', key: 'status', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    const rows = _.map(sortedRuns, function (jobRun) {
      return {
        component: (
          <Row
            key={jobRun.metadata.name}
            env={params.env}
            job={params.job}
            jobRun={jobRun}
          />
        )
      }
    })
    return (<Table columns={columns} rows={rows} />)
  }
}

@connect()
class Row extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    job: PropTypes.string,
    jobRun: PropTypes.object
  }

  deleteJobRun = () => {
    const { env, jobRun } = this.props
    this.props.dispatch(deleteJobRun(env, jobRun.metadata.name))
  }

  render () {
    const { env, job, jobRun } = this.props
    return (
      <tr>
        <td><Link to={`/${env}/jobs/${job}/runs/${jobRun.metadata.name}`}>{jobRun.metadata.name}</Link></td>
        <td>{jobRun.imageTag}</td>
        <td>{jobRun.startedAt}</td>
        <td>{jobRun.completedAt}</td>
        <td style={{textAlign: 'right'}}>{jobRun.statusName}</td>
        <td style={{textAlign: 'right'}}>
          <button type='button' className='btn btn-xs btn-danger' onClick={this.deleteJobRun}>Delete</button>
        </td>
      </tr>
    )
  }

}
