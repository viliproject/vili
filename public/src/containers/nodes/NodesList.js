import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { Button } from 'react-bootstrap'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { setNodeSchedulable } from '../../actions/nodes'

function mapStateToProps (state, ownProps) {
  const nodes = state.nodes.lookUpObjects(ownProps.params.env)
  return {
    nodes
  }
}

@connect(mapStateToProps)
export default class NodesList extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    nodes: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateNav('nodes'))
  }

  setNodeSchedulable = (node, action) => {
    this.props.dispatch(setNodeSchedulable(this.props.params.env, this.props.params.node, action))
  }

  render () {
    const { params, nodes } = this.props
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Nodes</li>
        </ol>
      </div>
    )
    const columns = [
      { title: 'Host', key: 'host' },
      { title: 'Instance Type', key: 'instance_type' },
      { title: 'Role', key: 'role' },
      { title: 'Capacity',
        key: 'capacity',
        subcolumns: [
          {title: 'CPU', key: 'cpu_capacity'},
          {title: 'Memory', key: 'memory_capacity'},
          {title: 'Pods', key: 'pods_capacity'}
        ]},
      { title: 'Versions',
        key: 'versions',
        subcolumns: [
          {title: 'CoreOS', key: 'os_version'},
          {title: 'Kubelet', key: 'kubelet_version'},
          {title: 'Proxy', key: 'proxy_version'}
        ]},
      { title: 'Created', key: 'created' },
      { title: 'Status', key: 'status' },
      { title: 'Actions', key: 'actions' }
    ]

    const rows = _.map(nodes, function (node) {
      var name = node.metadata.name
      var nodeStatuses = []
      if (node.status.conditions[0].status === 'Unknown') {
        nodeStatuses.push('NotReady')
      } else {
        nodeStatuses.push(node.status.conditions[0].type)
      }
      var actions
      if (node.spec.unschedulable === true) {
        actions = (
          <Button bsStyle='success' bsSize='xs'
            onClick={() => this.setNodeSchedulable(name, 'enable')}
          >
            Enable
          </Button>
        )
        nodeStatuses.push('Disabled')
      } else {
        actions = (
          <Button bsStyle='danger' bsSize='xs'
            onClick={() => this.setNodeSchedulable(name, 'disable')}
          >
            Disable
          </Button>
        )
      }

      return {
        host: <Link to={`/${params.env}/nodes/${name}`}>{name}</Link>,
        instance_type: node.metadata.labels['airware.io/instance-type'],
        role: node.metadata.labels['airware.io/role'],
        cpu_capacity: node.status.capacity.cpu,
        memory_capacity: node.memory,
        pods_capacity: node.status.capacity.pods,
        os_version: node.status.nodeInfo.osImage,
        kubelet_version: node.status.nodeInfo.kubeletVersion,
        proxy_version: node.status.nodeInfo.kubeProxyVersion,
        created: node.createdAt,
        status: nodeStatuses.join(','),
        actions: actions
      }
    })

    return (
      <div>
        {header}
        <Table columns={columns} rows={rows} />
      </div>
    )
  }

}
