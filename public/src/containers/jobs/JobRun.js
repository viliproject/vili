import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Alert } from 'react-bootstrap'
import _ from 'underscore'

import JobRunPod from '../../components/jobs/JobRunPod'
import { activateJobTab } from '../../actions/app'
import { makeLookUpObjectsByLabel } from '../../selectors'

function makeMapStateToProps () {
  const lookUpObjectsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { env, run } = ownProps.params
    const jobRun = state.jobRuns.lookUpObject(env, run)
    const pods = lookUpObjectsByLabel(state.pods, env, 'run', run)
    return {
      jobRun,
      pods
    }
  }
}

const dispatchProps = {
  activateJobTab
}

export class JobRun extends React.Component {
  static propTypes = {
    activateJobTab: PropTypes.func.isRequired,
    params: PropTypes.object,
    location: PropTypes.object,
    jobRun: PropTypes.object,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.activateJobTab('runs')
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
    const metadata = []
    metadata.push(<dt key='t-tag'>Tag</dt>)
    metadata.push(<dd key='d-tag'>{jobRun.imageTag}</dd>)
    if (jobRun.imageBranch) {
      metadata.push(<dt key='t-branch'>Branch</dt>)
      metadata.push(<dd key='d-branch'>{jobRun.imageBranch}</dd>)
    }
    metadata.push(<dt key='t-start-time'>Start Time</dt>)
    metadata.push(<dd key='d-start-time'>{jobRun.startedAt}</dd>)
    metadata.push(<dt key='t-completion-time'>Completion Time</dt>)
    metadata.push(<dd key='d-completion-time'>{jobRun.completedAt}</dd>)
    if (jobRun.startedBy) {
      metadata.push(<dt key='t-startedBy'>Started By</dt>)
      metadata.push(<dd key='d-startedBy'>{jobRun.startedBy}</dd>)
    }
    return metadata
  }

  render () {
    const { params, jobRun, pods } = this.props
    if (!jobRun) {
      return null
    }
    const podLogs = []
    pods.forEach((pod, podName) => {
      podLogs.push(<JobRunPod key={podName} env={params.env} podName={podName} />)
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

export default connect(makeMapStateToProps, dispatchProps)(JobRun)
