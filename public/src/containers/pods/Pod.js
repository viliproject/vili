import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import Loading from '../../components/Loading'
import PodLog from '../../components/PodLog'
import { activateNav } from '../../actions/app'
import { subPodLog, unsubPodLog } from '../../actions/pods'

function mapStateToProps (state, ownProps) {
  const { env, pod: podName } = ownProps.params
  const pod = state.pods.lookUpData(env, podName)
  return {
    pod
  }
}

const dispatchProps = {
  activateNav,
  subPodLog,
  unsubPodLog
}

@connect(mapStateToProps, dispatchProps)
export default class Pod extends React.Component {
  static propTypes = {
    activateNav: PropTypes.func.isRequired,
    subPodLog: PropTypes.func.isRequired,
    unsubPodLog: PropTypes.func.isRequired,
    params: PropTypes.object,
    location: PropTypes.object,
    pod: PropTypes.object
  }

  componentDidMount () {
    this.props.activateNav('pods')
    this.subData()
  }

  componentDidUpdate (prevProps) {
    const { params: { env, pod } } = this.props
    const { params: { env: prevEnv, pod: prevPod } } = prevProps
    if (env !== prevEnv || pod !== prevPod) {
      this.unsubData()
      this.subData()
    }
  }

  subData = () => {
    const { params, subPodLog } = this.props
    subPodLog(params.env, params.pod)
  }

  unsubData = () => {
    const { params, unsubPodLog } = this.props
    unsubPodLog(params.env, params.pod)
  }

  render () {
    const { params, pod: podData } = this.props
    const pod = podData && podData.get('object')
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li><Link to={`/${params.env}/pods`}>Pods</Link></li>
          <li className='active'>{params.pod}</li>
        </ol>
      </div>
    )
    if (!pod) {
      return (
        <div>
          {header}
          <Loading />
        </div>
      )
    }

    const metadata = [
      <dt key='title-ip'>IP</dt>,
      <dd key='data-ip'>{pod.getIn(['status', 'podIP'])}</dd>,
      <dt key='title-phase'>Phase</dt>,
      <dd key='data-phase'>{pod.getIn(['status', 'phase'])}</dd>,
      <dt key='title-node'>Node</dt>,
      (<dd key='data-node'>
        <Link to={`/${params.env}/nodes/${pod.getIn(['spec', 'nodeName'])}`}>{pod.getIn(['spec', 'nodeName'])}</Link>
      </dd>)
    ]
    if (pod.getLabel('app')) {
      metadata.push(<dt key='title-deployment'>Deployment</dt>)
      metadata.push(<dd key='data-deployment'>
        <Link to={`/${params.env}/deployments/${pod.getLabel('app')}`}>{pod.getLabel('app')}</Link>
      </dd>)
    }
    if (pod.getLabel('job')) {
      metadata.push(<dt key='title-job'>Job</dt>)
      metadata.push(<dd key='data-job'>
        <Link to={`/${params.env}/jobs/${pod.getLabel('job')}`}>{pod.getLabel('job')}</Link>
      </dd>)
    }
    if (pod.deployedBy) {
      metadata.push(<dt key='title-deployedBy'>Deployed By</dt>)
      metadata.push(<dd key='data-deployedBy'>{pod.deployedBy}</dd>)
    }
    return (
      <div>
        {header}
        <div>
          <h4>Metadata</h4>
          <dl className='dl-horizontal'>{metadata}</dl>
        </div>
        <PodLog log={podData.get('log')} />
      </div>
    )
  }

}
