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
    if (this.props.params !== prevProps.params) {
      this.unsubData()
      this.subData()
    }
  }

  subData = () => {
    const { params } = this.props
    this.props.subPodLog(params.env, params.pod)
  }

  unsubData = () => {
    const { params } = this.props
    this.props.unsubPodLog(params.env, params.pod)
  }

  render () {
    const { params, pod } = this.props
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li><Link to={`/${params.env}/pods`}>Pods</Link></li>
          <li className='active'>{params.pod}</li>
        </ol>
      </div>
    )
    if (!pod || !pod.object) {
      return (
        <div>
          {header}
          <Loading />
        </div>
      )
    }

    const metadata = [
      <dt key='title-ip'>IP</dt>,
      <dd key='data-ip'>{pod.object.status.podIP}</dd>,
      <dt key='title-phase'>Phase</dt>,
      <dd key='data-phase'>{pod.object.status.phase}</dd>,
      <dt key='title-node'>Node</dt>,
      (<dd key='data-node'>
        <Link to={`/${this.props.params.env}/nodes/${pod.object.spec.nodeName}`}>{pod.object.spec.nodeName}</Link>
      </dd>)
    ]
    if (pod.object.metadata.labels.app) {
      metadata.push(<dt key='title-deployment'>Deployment</dt>)
      metadata.push(<dd key='data-deployment'>
        <Link to={`/${this.props.params.env}/deployments/${pod.object.metadata.labels.app}`}>{pod.object.metadata.labels.app}</Link>
      </dd>)
    }
    if (pod.object.metadata.labels.job) {
      metadata.push(<dt key='title-job'>Job</dt>)
      metadata.push(<dd key='data-job'>
        <Link to={`/${this.props.params.env}/jobs/${pod.object.metadata.labels.job}`}>{pod.object.metadata.labels.job}</Link>
      </dd>)
    }
    if (pod.object.metadata.annotations['vili/deployedBy']) {
      metadata.push(<dt key='title-deployedBy'>Deployed By</dt>)
      metadata.push(<dd key='data-deployedBy'>{pod.object.metadata.annotations['vili/deployedBy']}</dd>)
    }
    return (
      <div>
        {header}
        <div>
          <h4>Metadata</h4>
          <dl className='dl-horizontal'>{metadata}</dl>
        </div>
        <PodLog log={pod.log} />
      </div>
    )
  }

}
