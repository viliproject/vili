import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import Table from '../../components/Table'
import { activateJobTab } from '../../actions/app'
import { deleteJobRun } from '../../actions/jobRuns'
import { makeLookUpObjectsByLabel } from '../../selectors'

function makeMapStateToProps () {
  const lookUpObjectsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { env, job: jobName } = ownProps.params
    const jobRuns = lookUpObjectsByLabel(state.jobRuns, env, 'job', jobName)
    return {
      jobRuns
    }
  }
}

const dispatchProps = {
  activateJobTab
}

export class JobRuns extends React.Component {
  static propTypes = {
    activateJobTab: PropTypes.func.isRequired,
    params: PropTypes.object,
    location: PropTypes.object,
    jobRuns: PropTypes.object
  }

  componentDidMount () {
    this.props.activateJobTab('runs')
  }

  render () {
    const { params, jobRuns } = this.props
    const columns = [
      {title: 'Run', key: 'run'},
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Start Time', key: 'startTime', style: {width: '180px'}},
      {title: 'Completion Time', key: 'completionTime', style: {width: '180px'}},
      {title: 'Status', key: 'status', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    const rows = []
    jobRuns
      .sortBy(r => -r.creationTimestamp)
      .forEach((jobRun) => {
        rows.push({
          component: (
            <Row
              key={jobRun.getIn(['metadata', 'name'])}
              env={params.env}
              job={params.job}
              jobRun={jobRun}
            />
          )
        })
      })
    return (<Table columns={columns} rows={rows} />)
  }
}

export default connect(makeMapStateToProps, dispatchProps)(JobRuns)

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
    this.props.dispatch(deleteJobRun(env, jobRun.getIn(['metadata', 'name'])))
  }

  render () {
    const { env, job, jobRun } = this.props
    const jobRunName = jobRun.getIn(['metadata', 'name'])
    return (
      <tr>
        <td><Link to={`/${env}/jobs/${job}/runs/${jobRunName}`}>{jobRunName}</Link></td>
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
