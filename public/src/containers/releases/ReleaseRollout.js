import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Panel, Alert } from 'react-bootstrap'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'

const tableColumns = {
  waveActions: [
    {title: 'Action', key: 'name'},
    {title: 'Branch', key: 'branch', style: {width: '200px', textAlign: 'right'}}
  ],
  waveJobs: [
    {title: 'Job', key: 'name'},
    {title: 'Branch', key: 'branch', style: {width: '200px', textAlign: 'right'}},
    {title: 'Tag', key: 'tag', style: {width: '200px', textAlign: 'right'}}
  ],
  waveApps: [
    {title: 'App', key: 'name'},
    {title: 'Branch', key: 'branch', style: {width: '200px', textAlign: 'right'}},
    {title: 'Tag', key: 'tag', style: {width: '200px', textAlign: 'right'}}
  ]
}

function mapStateToProps (state, ownProps) {
  const { env: envName, release: releaseName, rollout: rolloutID } = ownProps.params
  const env = _.findWhere(state.envs.toJS().envs, {name: envName})
  const release = state.releases.lookUpObject(envName, releaseName)
  const rollouts = release && release.envRollouts(envName) || []
  const rollout = _.findWhere(rollouts, {id: parseInt(rolloutID)})
  const rolloutWaves = rollout && rollout.waves || []
  return {
    env,
    release,
    rollout,
    rolloutWaves
  }
}

const dispatchProps = {
  activateNav
}

@connect(mapStateToProps, dispatchProps)
export default class ReleaseRollout extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    env: PropTypes.object,
    release: PropTypes.object,
    rollout: PropTypes.object,
    rolloutWaves: PropTypes.array,
    activateNav: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateNav('releases')
  }

  renderMetadata () {
    const { env, release, rollout } = this.props
    if (!env || !release || !rollout) {
      return null
    }
    const metadata = []

    switch (rollout.status) {
      case 'deployed':
        metadata.push(
          <Alert key='alert' bsStyle='success'>Release was rolled out <strong>{rollout.rolloutAtHumanize}</strong> by <strong>{rollout.rolloutBy}</strong></Alert>
        )
        break
      case 'deploying':
        metadata.push(
          <Alert key='alert' bsStyle='warning'>Release is rolling out, started by <strong>{rollout.rolloutBy}</strong></Alert>
        )
        break
      case 'failed':
        metadata.push(
          <Alert key='alert' bsStyle='danger'>Release rollout failed at <strong>{rollout.rolloutAtHumanize}</strong>, was started by <strong>{rollout.rolloutBy}</strong></Alert>
        )
        break
    }
    metadata.push(<h5 key='createdAt-title'>Created At</h5>)
    metadata.push(
      <div key='createdAt-value'>{release.createdAtHumanize}</div>
    )
    if (release.link) {
      metadata.push(<h5 key='link-title'>Link</h5>)
      metadata.push(
        <div key='link-value'><a href={release.link} target='_blank'>{release.link}</a></div>
      )
    }
    return metadata
  }

  renderWavePanels () {
    const { env, release, rolloutWaves } = this.props
    if (release) {
      const panels = _.map(release.waves, (wave, ix) => {
        const rolloutWave = rolloutWaves[ix] || {}
        return (
          <WavePanel
            key={ix}
            ix={ix}
            env={env.name}
            wave={wave}
            rolloutWave={rolloutWave}
          />
        )
      })
      return (
        <div>
          <h5>Waves</h5>
          {panels}
        </div>
      )
    }
    return null
  }

  render () {
    const { params } = this.props
    const header = [
      <div key='header' className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li><Link to={`/${params.env}/releases`}>Releases</Link></li>
          <li><Link to={`/${params.env}/releases/${params.release}`}>{params.release}</Link></li>
          <li className='active'>{`Rollout ${params.rollout}`}</li>
        </ol>
      </div>
    ]

    return (
      <div>
        {header}
        {this.renderMetadata()}
        {this.renderWavePanels()}
      </div>
    )
  }

}

class WavePanel extends React.Component {
  static propTypes = {
    env: PropTypes.string,
    ix: PropTypes.number,
    wave: PropTypes.object,
    rolloutWave: PropTypes.object,
    deployments: PropTypes.object,
    replicaSets: PropTypes.object,
    jobRuns: PropTypes.object
  }

  actionsTable () {
    const { targets } = this.props.wave
    const rows = _.map(
      _.filter(targets, (target) => target.type === 'action'),
      (target) => {
        return {
          name: target.name,
          branch: target.branch
        }
      })
    if (rows.length > 0) {
      return (
        <Table columns={tableColumns.waveActions} rows={rows} fill hover={false} />
      )
    }
    return null
  }

  jobsTable () {
    const { env } = this.props
    const { targets } = this.props.wave
    const rows = _.map(
      _.filter(targets, (target) => target.type === 'job'),
      (target) => {
        return {
          name: (<Link to={`/${env}/jobs/${target.name}`}>{target.name}</Link>),
          branch: target.branch,
          tag: target.tag,
          runAt: target.runAt
        }
      })
    if (rows.length > 0) {
      return (
        <Table columns={tableColumns.waveJobs} rows={rows} fill hover={false} />
      )
    }
    return null
  }

  appsTable () {
    const { env } = this.props
    const { targets } = this.props.wave
    const rows = _.map(
      _.filter(targets, (target) => target.type === 'app'),
      (target) => {
        return {
          name: (<Link to={`/${env}/deployments/${target.name}`}>{target.name}</Link>),
          branch: target.branch,
          tag: target.tag,
          deployedAt: target.deployedAt
        }
      })
    if (rows.length > 0) {
      return (
        <Table columns={tableColumns.waveApps} rows={rows} fill hover={false} />
      )
    }
    return null
  }

  get bsStyle () {
    const { status } = this.props.rolloutWave
    switch (status) {
      case 'deployed':
        return 'success'
      case 'deploying':
        return 'warning'
      case 'failed':
        return 'danger'
      default:
        return 'default'
    }
  }

  render () {
    const { ix } = this.props
    return (
      <Panel header={`Wave ${ix + 1}`} bsStyle={this.bsStyle}>
        {this.actionsTable()}
        {this.jobsTable()}
        {this.appsTable()}
      </Panel>
    )
  }

}
