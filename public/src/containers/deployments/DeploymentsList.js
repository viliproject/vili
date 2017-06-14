import PropTypes from 'prop-types'
/* global prompt */
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  const deployments = state.deployments.lookUpObjects(ownProps.params.env)
  const replicaSets = state.replicaSets.lookUpObjects(ownProps.params.env)
  return {
    env,
    deployments,
    replicaSets
  }
}

const dispatchProps = {
  activateNav
}

@connect(mapStateToProps, dispatchProps)
export default class DeploymentsList extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    env: PropTypes.object,
    deployments: PropTypes.object,
    replicaSets: PropTypes.object,
    activateNav: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateNav('deployments')
  }

  render () {
    const { params, env, deployments, replicaSets } = this.props

    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Deployments</li>
        </ol>
      </div>
    )

    const columns = [
      {title: 'Name', key: 'name'},
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Replicas', key: 'replicas', style: {width: '100px', textAlign: 'right'}},
      {title: 'Deployed', key: 'deployedAt', style: {width: '200px', textAlign: 'right'}}
    ]

    const rows = _.map(env.deployments, (deploymentName) => {
      const deployment = deployments[deploymentName]
      const replicaSet = deployment && _.find(replicaSets, (rs) => {
        return rs.hasLabel('app', deploymentName) && rs.revision === deployment.revision
      })
      return {
        name: (<Link to={`/${env.name}/deployments/${deploymentName}`}>{deploymentName}</Link>),
        tag: replicaSet && replicaSet.imageTag,
        replicas: replicaSet && (replicaSet.status.replicas + '/' + replicaSet.spec.replicas),
        deployedAt: replicaSet && replicaSet.deployedAt
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
