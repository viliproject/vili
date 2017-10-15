import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { makeLookUpObjects } from '../../selectors'

function makeMapStateToProps () {
  const lookUpDeploymentObjects = makeLookUpObjects()
  const lookUpReplicaSetObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { env: envName } = ownProps.params
    const env = state.envs.getIn(['envs', envName])
    const deployments = lookUpDeploymentObjects(state.deployments, env.name)
    const replicaSets = lookUpReplicaSetObjects(state.replicaSets, env.name)
    return {
      env,
      deployments,
      replicaSets
    }
  }
}

const dispatchProps = {
  activateNav
}

export class DeploymentsList extends React.Component {
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

    const rows = []
    env.deployments.forEach((deploymentName) => {
      const deployment = deployments.find((d) => d.getIn(['metadata', 'name']) === deploymentName)
      const replicaSet = deployment && replicaSets
        .filter(x => x.hasLabel('app', deploymentName) && x.revision === deployment.revision)
        .sortBy(x => -x.creationTimestamp)
        .first()
      rows.push({
        name: (<Link to={`/${env.name}/deployments/${deploymentName}`}>{deploymentName}</Link>),
        tag: deployment && deployment.imageTag,
        replicas: replicaSet && `${replicaSet.getIn(['status', 'replicas'])}/${replicaSet.getIn(['spec', 'replicas'])}`,
        deployedAt: replicaSet && replicaSet.deployedAt
      })
    })

    return (
      <div>
        {header}
        <Table columns={columns} rows={rows} />
      </div>
    )
  }
}

export default connect(makeMapStateToProps, dispatchProps)(DeploymentsList)
