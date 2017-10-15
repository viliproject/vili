import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { makeLookUpObjectsByNodeName } from '../../selectors'

function makeMapStateToProps () {
  const lookUpPodsByNodeName = makeLookUpObjectsByNodeName()
  return (state, ownProps) => {
    const { env, node: nodeName } = ownProps.params
    const node = state.nodes.lookUpData(env, nodeName)
    const pods = lookUpPodsByNodeName(state.pods, env, nodeName)
    return {
      node,
      pods
    }
  }
}

const dispatchProps = {
  activateNav
}

export class Node extends React.Component {
  componentDidMount () {
    this.props.activateNav('nodes')
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
      {title: 'Pod IP', key: 'podIP'},
      {title: 'Created', key: 'created'},
      {title: 'Phase', key: 'phase'}
    ])

    const rows = []
    pods.forEach((pod) => {
      var app = null
      if (pod.getLabel('app')) {
        app = <Link to={`/${params.env}/deployments/${pod.getLabel('app')}`}>{pod.getLabel('app')}</Link>
      }
      rows.push({
        name: <Link to={`/${params.env}/pods/${pod.getIn(['metadata', 'name'])}`}>{pod.getIn(['metadata', 'name'])}</Link>,
        app: app,
        phase: pod.getIn(['status', 'phase']),
        podIP: pod.getIn(['status', 'podIP']),
        created: pod.createdAt
      })
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

Node.propTypes = {
  activateNav: PropTypes.func,
  params: PropTypes.object,
  location: PropTypes.object,
  node: PropTypes.object,
  pods: PropTypes.object
}

export default connect(makeMapStateToProps, dispatchProps)(Node)
