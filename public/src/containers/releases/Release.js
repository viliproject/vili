import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Button, ButtonToolbar, Panel } from 'react-bootstrap'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { deployRelease } from '../../actions/releases'

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
  const { env: envName, release: releaseName } = ownProps.params
  const env = _.findWhere(state.envs.toJS().envs, {name: envName})
  const release = state.releases.lookUpObject(envName, releaseName)
  const rollouts = release && release.envRollouts(envName) || []
  return {
    env,
    release,
    rollouts
  }
}

const dispatchProps = {
  activateNav,
  deployRelease
}

@connect(mapStateToProps, dispatchProps)
export default class Release extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    env: PropTypes.object,
    release: PropTypes.object,
    rollouts: PropTypes.array,
    activateNav: PropTypes.func.isRequired,
    deployRelease: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateNav('releases')
  }

  deployRelease = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { deployRelease, env, release } = this.props
    deployRelease(env.name, release.name)
  }

  renderActions () {
    const buttons = []
    buttons.push(
      <Button key='deploy' onClick={this.deployRelease} bsStyle='primary' bsSize='small'>Deploy</Button>
    )
    return (
      <ButtonToolbar key='toolbar' className='pull-right'>
        {buttons}
      </ButtonToolbar>
    )
  }

  renderMetadata () {
    const { env, release, rollouts } = this.props
    if (!env || !release) {
      return null
    }

    const metadata = []
    if (env.approvedFromEnv) {
      metadata.push(<h5 key='approved-env-title'>Approved From</h5>)
      metadata.push(<div key='approved-env-value'>{env.approvedFromEnv}</div>)
    }

    if (release.link) {
      metadata.push(<h5 key='link-title'>Link</h5>)
      metadata.push(
        <div key='link-value'><a href={release.link} target='_blank'>{release.link}</a></div>
      )
    }

    metadata.push(<h5 key='approvedBy-title'>Approved By</h5>)
    metadata.push(
      <div key='approvedBy-value'>{release.createdBy}</div>
    )
    metadata.push(<h5 key='createdAt-title'>Created At</h5>)
    metadata.push(
      <div key='createdAt-value'>{release.createdAtHumanize}</div>
    )

    if (rollouts.length > 0) {
      metadata.push(<h5 key='rollouts-title'>Rollouts</h5>)

      const columns = [
        {title: 'ID', key: 'id', style: {width: '50px'}},
        {title: 'Rollout At', key: 'rolloutAtHumanize'},
        {title: 'Rollout By', key: 'rolloutBy', style: {width: '200px', textAlign: 'right'}},
        {title: 'Status', key: 'status', style: {width: '200px', textAlign: 'right'}}
      ]
      const rows = _.map(rollouts, (rollout) => {
        return {
          id: (<Link to={`/${env.name}/releases/${release.name}/rollouts/${rollout.id}`}>{rollout.id}</Link>),
          rolloutAtHumanize: rollout.rolloutAtHumanize,
          rolloutBy: rollout.rolloutBy,
          status: rollout.status
        }
      })
      metadata.push(
        <Table key='rollouts-value' columns={columns} rows={rows} />
      )
    }
    return metadata
  }

  renderWavePanels () {
    const { env, release } = this.props
    if (release) {
      const panels = _.map(release.waves, (wave, ix) => {
        return (
          <WavePanel
            key={ix}
            ix={ix}
            env={env.name}
            wave={wave}
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
    const header = (
      <div key='header' className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li><Link to={`/${params.env}/releases`}>Releases</Link></li>
          <li className='active'>{params.release}</li>
        </ol>
        {this.renderActions()}
      </div>
    )

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

  render () {
    const { ix } = this.props
    return (
      <Panel header={`Wave ${ix + 1}`}>
        {this.actionsTable()}
        {this.jobsTable()}
        {this.appsTable()}
      </Panel>
    )
  }

}
