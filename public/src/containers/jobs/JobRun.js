import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Alert } from 'react-bootstrap'
import _ from 'underscore'

import JobRunPod from '../../components/jobs/JobRunPod'
import { activateJobTab } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const { env, run } = ownProps.params
  const jobRun = state.jobRuns.lookUpObject(env, run)
  const pods = state.pods.lookUpObjectsByFunc(env, (obj) => {
    return obj.hasLabel('run', run)
  })
  return {
    jobRun,
    pods
  }
}

@connect(mapStateToProps)
export default class JobRun extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    jobRun: PropTypes.object,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateJobTab('runs'))
  }

  renderBanner () {
    const { jobRun } = this.props
    var banner = null
    switch (jobRun.statusName) {
      case 'Failed':
        banner = (<Alert bsStyle='danger'>Failed</Alert>)
        break
      case 'Complete':
        banner = (<Alert bsStyle='success'>Complete</Alert>)
        break
    }
    return banner
  }

  renderMetadata () {
    const { jobRun } = this.props
    const metadata = [
      <dt key='title-tag'>Tag</dt>,
      <dd key='data-tag'>{jobRun.imageTag}</dd>,
      <dt key='title-start-time'>Start Time</dt>,
      <dd key='data-start-time'>{jobRun.startedAt}</dd>,
      <dt key='title-completion-time'>Completion Time</dt>,
      <dd key='data-completion-time'>{jobRun.completedAt}</dd>
    ]
    return metadata
  }

  render () {
    const { params, jobRun, pods } = this.props
    if (!jobRun) {
      return null
    }
    const podLogs = _.map(pods, (pod, podName) => {
      return <JobRunPod key={podName} env={params.env} podName={podName} />
    })
    if (podLogs.length > 0) {
      podLogs.splice(0, 0, (<h3 key='header'>Pods</h3>))
    }
    return (
      <div>
        <div>
          {this.renderBanner()}
          <dl className='dl-horizontal'>{this.renderMetadata()}</dl>
        </div>
        {podLogs}
      </div>
    )
  }
}
