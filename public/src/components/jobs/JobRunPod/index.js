import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import PodLog from '../../../components/PodLog'
import { subPodLog, unsubPodLog } from '../../../actions/pods'

function mapStateToProps (state, ownProps) {
  const pod = state.pods.lookUpData(ownProps.env, ownProps.podName)
  return {
    log: pod.log
  }
}

@connect(mapStateToProps)
export default class JobRunPod extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
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
    const { env, podName } = this.props
    this.props.dispatch(subPodLog(env, podName))
  }

  unsubData = () => {
    const { env, podName } = this.props
    this.props.dispatch(unsubPodLog(env, podName))
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
