import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import PodLog from '../../../components/PodLog'
import { subPodLog, unsubPodLog } from '../../../actions/pods'

function mapStateToProps (state, ownProps) {
  const { env, podName } = ownProps
  const pod = state.pods.lookUpData(env, podName)
  return {
    log: pod.get('log')
  }
}

const dispatchProps = {
  subPodLog,
  unsubPodLog
}

export class JobRunPod extends React.Component {
  static propTypes = {
    subPodLog: PropTypes.func.isRequired,
    unsubPodLog: PropTypes.func.isRequired,
    env: PropTypes.string,
    podName: PropTypes.string,
    log: PropTypes.string
  }

  componentDidMount () {
    this.subData()
  }

  componentWillUnmount () {
    this.unsubData()
  }

  subData = () => {
    const { env, podName, subPodLog } = this.props
    subPodLog(env, podName)
  }

  unsubData = () => {
    const { env, podName, unsubPodLog } = this.props
    unsubPodLog(env, podName)
  }

  render () {
    const { env, podName, log } = this.props
    return (
      <div key={podName}>
        <h4>
          <Link to={`/${env}/pods/${podName}`}>{podName}</Link>
        </h4>
        <PodLog log={log} />
      </div>
    )
  }
}

export default connect(mapStateToProps, dispatchProps)(JobRunPod)
