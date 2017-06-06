import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const { env, node: nodeName } = ownProps.params
  const node = state.nodes.lookUpData(env, nodeName)
  const pods = state.pods.lookUpObjectsByFunc(env, (obj) => {
    return obj.spec.nodeName === nodeName
  })
  return {
    node,
    pods
  }
}

@connect(mapStateToProps)
export default class Node extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    node: PropTypes.object,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateNav('nodes'))
  }

  render () {
    const { params, pods } = this.props
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li><Link to={`/${params.env}/nodes`}>Nodes</Link></li>
          <li className='active'>{params.node}</li>
        </ol>
      </div>
    )
    const columns = _.union([
      {title: 'Name', key: 'name'},
      {title: 'App', key: 'app'},
      {title: 'Pod IP', key: 'pod_ip'},
      {title: 'Created', key: 'created'},
      {title: 'Phase', key: 'phase'}
    ])

    const rows = _.map(pods, function (pod) {
      var app = null
      if (pod.metadata.labels && pod.metadata.labels.app) {
        app = <Link to={`/${params.env}/deployments/${pod.metadata.labels.app}`}>{pod.metadata.labels.app}</Link>
      }
      return {
        name: <Link to={`/${params.env}/pods/${pod.metadata.name}`}>{pod.metadata.name}</Link>,
        app: app,
        phase: pod.status.phase,
        pod_ip: pod.status.podIP,
        created: pod.createdAt
      }
    })

    return (
      <div>
        {header}
        <div>
          <h3>Pods</h3>
          <Table columns={columns} rows={rows} />
        </div>
      </div>
    )
  }

}
