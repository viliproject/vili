import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { Button } from 'react-bootstrap'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { deletePod } from '../../actions/pods'

function mapStateToProps (state, ownProps) {
  const pods = state.pods.lookUpObjects(ownProps.params.env)
  return {
    pods
  }
}

@connect(mapStateToProps)
export default class PodsList extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateNav('pods'))
  }

  render () {
    const { params, pods } = this.props
    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Pods</li>
        </ol>
      </div>
    )

    const columns = [
      {title: 'Name', key: 'name'},
      {title: 'Deployment/Job', key: 'deployment-job'},
      {title: 'Node', key: 'node'},
      {title: 'Phase', key: 'phase'},
      {title: 'Ready', key: 'ready'},
      {title: 'Created', key: 'created'},
      {title: 'Actions', key: 'actions'}
    ]

    const rows = _.map(pods, function (pod, key) {
      return {
        component: (
          <Row key={key}
            env={params.env}
            pod={pod}
          />
        ),
        key: key
      }
    })
    const sortedRows = _.sortBy(rows, 'key')

    return (
      <div>
        {header}
        <Table columns={columns} rows={sortedRows} />
      </div>
    )
  }

}

@connect()
class Row extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    pod: PropTypes.object
  }

  get nameLink () {
    const { env, pod } = this.props
    return (
      <Link to={`/${env}/pods/${pod.metadata.name}`}>{pod.metadata.name}</Link>
    )
  }

  get deploymentJobLink () {
    const { env, pod } = this.props
    if (pod.metadata.labels.app) {
      return (
        <Link to={`/${env}/deployments/${pod.metadata.labels.app}`}>{pod.metadata.labels.app}</Link>
      )
    } else if (pod.metadata.labels.job) {
      return (
        <Link to={`/${env}/jobs/${pod.metadata.labels.job}`}>{pod.metadata.labels.job}</Link>
      )
    }
  }

  get nodeLink () {
    const { env, pod } = this.props
    return (
      <Link to={`/${env}/nodes/${pod.spec.nodeName}`}>{pod.spec.nodeName}</Link>
    )
  }

  deletePod = () => {
    const { env, pod } = this.props
    this.props.dispatch(deletePod(env, pod.metadata.name))
  }

  render () {
    const { pod } = this.props
    return (
      <tr>
        <td>{this.nameLink}</td>
        <td>{this.deploymentJobLink}</td>
        <td>{this.nodeLink}</td>
        <td>{pod.status.phase}</td>
        <td>{pod.isReady ? String.fromCharCode('10003') : ''}</td>
        <td>{pod.createdAt}</td>
        <td style={{textAlign: 'right'}}><Button onClick={this.deletePod} bsStyle='danger' bsSize='xs'>Delete</Button></td>
      </tr>
    )
  }

}
