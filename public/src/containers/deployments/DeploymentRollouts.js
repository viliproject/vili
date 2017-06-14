import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { Panel, ButtonToolbar, Button, Badge } from 'react-bootstrap'
import _ from 'underscore'

import Table from '../../components/Table'
import Loading from '../../components/Loading'
import { activateDeploymentTab } from '../../actions/app'
import { resumeDeployment, pauseDeployment, scaleDeployment, rollbackToRevision } from '../../actions/deployments'
import { deletePod } from '../../actions/pods'

function mapStateToProps (state, ownProps) {
  const { env, deployment: deploymentName } = ownProps.params
  const deployment = state.deployments.lookUpData(env, deploymentName)
  const unsortedReplicaSets = state.replicaSets.lookUpObjectsByFunc(env, (obj) => {
    return obj.hasLabel('app', deploymentName)
  })
  const replicaSets = _.sortBy(unsortedReplicaSets, x => -x.revision)
  const pods = state.pods.lookUpObjectsByFunc(env, (obj) => {
    return obj.hasLabel('app', deploymentName)
  })
  return {
    deployment,
    replicaSets,
    pods
  }
}

@connect(mapStateToProps)
export default class DeploymentRollouts extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    deployment: PropTypes.object,
    replicaSets: PropTypes.array,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateDeploymentTab('rollouts'))
  }

  resumeDeployment = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { params } = this.props
    this.props.dispatch(resumeDeployment(params.env, params.deployment))
  }

  pauseDeployment = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { params } = this.props
    this.props.dispatch(pauseDeployment(params.env, params.deployment))
  }

  scaleDeployment = () => {
    var replicas = prompt('Enter the number of replicas to scale to')
    if (!replicas) {
      return
    }
    replicas = parseInt(replicas)
    if (_.isNaN(replicas)) {
      return
    }
    const { params } = this.props
    this.props.dispatch(scaleDeployment(params.env, params.deployment, replicas))
  }

  renderHeader () {
    const buttons = []
    const { deployment } = this.props
    if (deployment.object.spec.paused) {
      buttons.push(
        <Button key='resume' bsStyle='success' bsSize='small' onClick={this.resumeDeployment}>Resume</Button>
      )
    } else {
      buttons.push(
        <Button key='pause' bsStyle='warning' bsSize='small' onClick={this.pauseDeployment}>Pause</Button>
      )
    }
    buttons.push(
      <Button key='scale' bsStyle='info' bsSize='small' onClick={this.scaleDeployment}>Scale</Button>
    )
    return (
      <div className='clearfix' style={{marginBottom: '10px'}}>
        <ButtonToolbar className='pull-right'>{buttons}</ButtonToolbar>
      </div>
    )
  }

  renderPodPanels () {
    const { params, deployment, replicaSets, pods } = this.props
    return _.map(replicaSets, (replicaSet) => {
      const replicaSetPods = getReplicaSetPods(replicaSet, pods)
      if (replicaSetPods.length === 0 && deployment.object.revision !== replicaSet.revision) {
        return null
      }
      return (
        <RolloutPanel
          key={replicaSet.metadata.name}
          env={params.env}
          deployment={deployment.object}
          replicaSet={replicaSet}
          pods={replicaSetPods}
        />
      )
    })
  }

  renderHistoryTable () {
    const { params, deployment, replicaSets, pods } = this.props
    const columns = [
      {title: 'Revision', key: 'revision', style: {width: '90px'}},
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Branch', key: 'branch'},
      {title: 'Rollout Time', key: 'time', style: {textAlign: 'right'}},
      {title: 'Deployed By', key: 'deployedBy', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    const rows = _.map(replicaSets, (replicaSet) => {
      const replicaSetPods = getReplicaSetPods(replicaSet, pods)
      if (replicaSetPods.length > 0 || deployment.object.revision === replicaSet.revision) {
        return null
      }
      return {
        component: (
          <HistoryRow
            key={replicaSet.metadata.name}
            env={params.env}
            deployment={params.deployment}
            replicaSet={replicaSet}
          />
        )
      }
    })
    return (
      <div>
        <h4>Rollout History</h4>
        <Table columns={columns} rows={rows} />
      </div>
    )
  }

  render () {
    const { deployment } = this.props
    if (!deployment || !deployment.object) {
      return (<Loading />)
    }
    return (
      <div>
        {this.renderHeader()}
        {this.renderPodPanels()}
        {this.renderHistoryTable()}
      </div>
    )
  }

}

@connect()
class RolloutPanel extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    deployment: PropTypes.object,
    replicaSet: PropTypes.object,
    pods: PropTypes.array
  }

  deletePod = (event, pod) => {
    event.target.setAttribute('disabled', 'disabled')
    this.props.dispatch(deletePod(this.props.env, pod))
  }

  renderReplicaSetMetadata () {
    const { replicaSet } = this.props
    const metadata = []
    metadata.push(<dt key='t-tag'>Tag</dt>)
    metadata.push(<dd key='d-tag'>{replicaSet.imageTag}</dd>)
    if (replicaSet.imageBranch) {
      metadata.push(<dt key='t-branch'>Branch</dt>)
      metadata.push(<dd key='d-branch'>{replicaSet.imageBranch}</dd>)
    }
    metadata.push(<dt key='t-time'>Time</dt>)
    metadata.push(<dd key='d-time'>{replicaSet.deployedAt}</dd>)
    if (replicaSet.deployedBy) {
      metadata.push(<dt key='t-deployedBy'>Deployed By</dt>)
      metadata.push(<dd key='d-deployedBy'>{replicaSet.deployedBy}</dd>)
    }
    return (
      <div>
        <dl>
          {metadata}
        </dl>
      </div>
    )
  }

  renderPodsTable () {
    const self = this
    const { env, pods } = this.props
    if (pods.length === 0) {
      return null
    }
    const columns = [
      {title: 'Pod', key: 'name'},
      {title: 'Host', key: 'host'},
      {title: 'Phase', key: 'phase'},
      {title: 'Ready', key: 'ready'},
      {title: 'Pod IP', key: 'podIP'},
      {title: 'Created', key: 'created'},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    const rows = _.map(pods, (pod) => {
      const nameLink = (<Link to={`/${env}/pods/${pod.metadata.name}`}>{pod.metadata.name}</Link>)
      const hostLink = (<Link to={`/${env}/nodes/${pod.spec.nodeName}`}>{pod.spec.nodeName}</Link>)
      var actions = (
        <Button
          bsStyle='danger'
          bsSize='xs'
          onClick={(event) => self.deletePod(event, pod.metadata.name)}
        >
          Delete
        </Button>
      )
      return {
        key: pod.metadata.name,
        name: nameLink,
        host: hostLink,
        phase: pod.status.phase,
        ready: pod.isReady ? String.fromCharCode('10003') : '',
        pod_ip: pod.status.podIP,
        created: pod.createdAt,
        actions
      }
    })

    return (<Table columns={columns} rows={rows} fill />)
  }

  render () {
    const { deployment, replicaSet, pods } = this.props
    var bsStyle = ''
    if (replicaSet.revision === deployment.revision) {
      bsStyle = 'success'
    } else {
      bsStyle = 'warning'
    }

    const header = (
      <div>
        <strong>{replicaSet.revision}</strong>
        <Badge bsStyle={bsStyle} pullRight>{pods.length}/{replicaSet.status.replicas}</Badge>
      </div>
    )
    return (
      <Panel
        key={replicaSet.metadata.name}
        header={header}
        bsStyle={bsStyle}
      >
        {this.renderReplicaSetMetadata()}
        {this.renderPodsTable()}
      </Panel>
    )
  }
}

@connect()
class HistoryRow extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    deployment: PropTypes.string,
    replicaSet: PropTypes.object
  }

  rollbackTo = (event) => {
    const { env, deployment, replicaSet } = this.props
    event.target.setAttribute('disabled', 'disabled')
    this.props.dispatch(rollbackToRevision(env, deployment, replicaSet.revision))
  }

  render () {
    const { replicaSet } = this.props
    return (
      <tr>
        <td>{replicaSet.revision}</td>
        <td>{replicaSet.imageTag}</td>
        <td>{replicaSet.metadata.annotations['vili/branch']}</td>
        <td style={{textAlign: 'right'}}>{replicaSet.deployedAt}</td>
        <td style={{textAlign: 'right'}}>{replicaSet.metadata.annotations['vili/deployedBy']}</td>
        <td style={{textAlign: 'right'}}>
          <Button bsStyle='danger' bsSize='xs' onClick={this.rollbackTo}>Rollback To</Button>
        </td>
      </tr>
    )
  }
}

function getReplicaSetPods (replicaSet, pods) {
  const generateName = replicaSet.metadata.name + '-'
  return _.filter(pods, (pod) => {
    return pod.metadata.generateName === generateName
  })
}
