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
import { makeLookUpObjectsByLabel } from '../../selectors'

function makeMapStateToProps () {
  const lookUpReplicaSetsByLabel = makeLookUpObjectsByLabel()
  const lookUpPodsByLabel = makeLookUpObjectsByLabel()
  return (state, ownProps) => {
    const { env, deployment: deploymentName } = ownProps.params
    const deployment = state.deployments.lookUpData(env, deploymentName)
    const replicaSets = lookUpReplicaSetsByLabel(state.replicaSets, env, 'app', deploymentName)
    const pods = lookUpPodsByLabel(state.pods, env, 'app', deploymentName)
    return {
      deployment,
      replicaSets,
      pods
    }
  }
}

const dispatchProps = {
  activateDeploymentTab,
  resumeDeployment,
  pauseDeployment,
  scaleDeployment
}

export class DeploymentRollouts extends React.Component {
  static propTypes = {
    activateDeploymentTab: PropTypes.func,
    resumeDeployment: PropTypes.func,
    pauseDeployment: PropTypes.func,
    scaleDeployment: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    deployment: PropTypes.object,
    replicaSets: PropTypes.object,
    pods: PropTypes.object
  }

  componentDidMount () {
    this.props.activateDeploymentTab('rollouts')
  }

  resumeDeployment = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { params, resumeDeployment } = this.props
    resumeDeployment(params.env, params.deployment)
  }

  pauseDeployment = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { params, pauseDeployment } = this.props
    pauseDeployment(params.env, params.deployment)
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
    const { params, scaleDeployment } = this.props
    scaleDeployment(params.env, params.deployment, replicas)
  }

  renderHeader () {
    const buttons = []
    const { deployment } = this.props
    if (deployment.getIn(['object', 'spec', 'paused'])) {
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

  renderPodPanels (replicaSets, activeReplicaSet) {
    const { params, deployment, pods } = this.props
    const podPanels = []
    replicaSets.forEach((replicaSet) => {
      const replicaSetPods = getReplicaSetPods(replicaSet, pods)
      if (replicaSetPods.isEmpty() && activeReplicaSet !== replicaSet) {
        return null
      }
      podPanels.push(
        <RolloutPanel
          key={replicaSet.getIn(['metadata', 'name'])}
          env={params.env}
          deployment={deployment.get('object')}
          replicaSet={replicaSet}
          pods={replicaSetPods}
        />
      )
    })
    return podPanels
  }

  renderHistoryTable (replicaSets, activeReplicaSet) {
    const { params, pods } = this.props
    const columns = [
      {title: 'Revision', key: 'revision', style: {width: '90px'}},
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Branch', key: 'branch'},
      {title: 'Rollout Time', key: 'time', style: {textAlign: 'right'}},
      {title: 'Deployed By', key: 'deployedBy', style: {textAlign: 'right'}},
      {title: 'Actions', key: 'actions', style: {textAlign: 'right'}}
    ]

    const rows = []
    replicaSets.forEach((replicaSet) => {
      const replicaSetPods = getReplicaSetPods(replicaSet, pods)
      if (!replicaSetPods.isEmpty() || activeReplicaSet === replicaSet) {
        return null
      }
      rows.push({
        component: (
          <HistoryRow
            key={replicaSet.getIn(['metadata', 'name'])}
            env={params.env}
            deployment={params.deployment}
            replicaSet={replicaSet}
          />
        )
      })
    })
    return (
      <div>
        <h4>Rollout History</h4>
        <Table columns={columns} rows={rows} />
      </div>
    )
  }

  render () {
    const { deployment, replicaSets } = this.props
    if (!deployment || !deployment.get('object')) {
      return (<Loading />)
    }
    const sortedReplicaSets = replicaSets
      .sort((a, b) => {
        if (a.revision > b.revision || a.creationTimestamp < b.creationTimestamp) {
          return 1
        }
        if (a.revision === b.revision && a.creationTimestamp === b.creationTimestamp) {
          return 0
        }
        return -1
      })
    const activeReplicaSet = sortedReplicaSets
      .filter(x => x.revision === deployment.get('object').revision)
      .first()

    return (
      <div>
        {this.renderHeader()}
        {this.renderPodPanels(sortedReplicaSets, activeReplicaSet)}
        {this.renderHistoryTable(sortedReplicaSets, activeReplicaSet)}
      </div>
    )
  }

}

export default connect(makeMapStateToProps, dispatchProps)(DeploymentRollouts)

@connect()
class RolloutPanel extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    deployment: PropTypes.object,
    replicaSet: PropTypes.object,
    pods: PropTypes.object
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
    if (pods.isEmpty()) {
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

    const rows = []
    pods.forEach((pod) => {
      const nameLink = (<Link to={`/${env}/pods/${pod.getIn(['metadata', 'name'])}`}>{pod.getIn(['metadata', 'name'])}</Link>)
      const hostLink = (<Link to={`/${env}/nodes/${pod.getIn(['spec', 'nodeName'])}`}>{pod.getIn(['spec', 'nodeName'])}</Link>)
      var actions = (
        <Button
          bsStyle='danger'
          bsSize='xs'
          onClick={(event) => self.deletePod(event, pod.getIn(['metadata', 'name']))}
        >
          Delete
        </Button>
      )
      rows.push({
        key: pod.getIn(['metadata', 'name']),
        name: nameLink,
        host: hostLink,
        phase: pod.getIn(['status', 'phase']),
        ready: pod.isReady ? String.fromCharCode('10003') : '',
        podIP: pod.getIn(['status', 'podIP']),
        created: pod.createdAt,
        actions
      })
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
        <Badge bsStyle={bsStyle} pullRight>{pods.size}/{replicaSet.getIn(['status', 'replicas'])}</Badge>
      </div>
    )
    return (
      <Panel
        key={replicaSet.getIn(['metadata', 'name'])}
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
        <td>{replicaSet.imageBranch}</td>
        <td style={{textAlign: 'right'}}>{replicaSet.deployedAt}</td>
        <td style={{textAlign: 'right'}}>{replicaSet.deployedBy}</td>
        <td style={{textAlign: 'right'}}>
          <Button bsStyle='danger' bsSize='xs' onClick={this.rollbackTo}>Rollback To</Button>
        </td>
      </tr>
    )
  }
}

function getReplicaSetPods (replicaSet, pods) {
  const generateName = replicaSet.getIn(['metadata', 'name']) + '-'
  return pods.filter((pod) => pod.getIn(['metadata', 'generateName']) === generateName)
}
