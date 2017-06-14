import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Button, ButtonToolbar, Panel, FormGroup, FormControl, ControlLabel, HelpBlock } from 'react-bootstrap'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { getReleaseSpec, createRelease } from '../../actions/releases'

const tableColumns = {
  waveActions: [
    {title: 'Action', key: 'name'},
    {title: 'Branch', key: 'branch', style: {width: '200px', textAlign: 'right'}}
  ],
  waveJobs: [
    {title: 'Job', key: 'name'},
    {title: 'Branch', key: 'branch', style: {width: '200px'}},
    {title: 'Tag', key: 'tag', style: {width: '200px'}},
    {title: 'Run At', key: 'runAt', style: {width: '200px', textAlign: 'right'}}
  ],
  waveApps: [
    {title: 'App', key: 'name'},
    {title: 'Branch', key: 'branch', style: {width: '200px'}},
    {title: 'Tag', key: 'tag', style: {width: '200px'}},
    {title: 'Deployed At', key: 'deployedAt', style: {width: '200px', textAlign: 'right'}}
  ]
}

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  const releaseEnv = state.releases.lookUp(ownProps.params.env)
  const deployments = state.deployments.lookUpObjects(ownProps.params.env)
  const replicaSets = state.replicaSets.lookUpObjects(ownProps.params.env)
  const jobRuns = state.jobRuns.lookUpObjects(ownProps.params.env)
  var spec
  if (releaseEnv && releaseEnv.spec) {
    spec = JSON.parse(JSON.stringify(releaseEnv.spec))
    _.each(spec.waves, (wave, ix) => {
      _.each(wave.targets, (target) => {
        updateTargetVersion(target, env, deployments, replicaSets, jobRuns)
      })
    })
  }
  return {
    env,
    spec,
    deployments,
    replicaSets,
    jobRuns
  }
}

const dispatchProps = {
  activateNav,
  getReleaseSpec,
  createRelease
}

function updateTargetVersion (target, env, deployments, replicaSets, jobRuns) {
  switch (target.type) {
    case 'action':
      target.branch = env.branch
      return
    case 'job':
      const runs = _.sortBy(
        _.filter(jobRuns, x => x.hasLabel('job', target.name)),
        x => -x.creationTimestamp
      )
      if (runs.length > 0) {
        target.tag = runs[0].imageTag
        target.branch = runs[0].imageBranch || env.branch
        target.runAt = runs[0].runAt
      }
      return
    case 'app':
      const deployment = deployments && deployments[target.name]
      const replicaSet = deployment && _.find(replicaSets, (rs) => {
        return rs.hasLabel('app', target.name) && rs.revision === deployment.revision
      })
      if (replicaSet) {
        target.tag = replicaSet.imageTag
        target.branch = replicaSet.imageBranch || env.branch
        target.deployedAt = replicaSet.deployedAt
      }
      return
  }
  return
}

@connect(mapStateToProps, dispatchProps)
export default class ReleaseCreate extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    env: PropTypes.object,
    spec: PropTypes.object,
    deployments: PropTypes.object,
    replicaSets: PropTypes.object,
    jobRuns: PropTypes.object,
    activateNav: PropTypes.func.isRequired,
    getReleaseSpec: PropTypes.func.isRequired,
    createRelease: PropTypes.func.isRequired
  }

  constructor (props) {
    super(props)

    this.state = {
      releaseName: '',
      releaseNameValidation: 'warning',
      releaseNameHelp: 'Release name cannot be empty',
      releaseLink: ''
    }
  }

  componentDidMount () {
    this.props.activateNav('releases')
    this.props.getReleaseSpec(this.props.params.env)
  }

  handleNameChange = (e) => {
    const releaseName = e.target.value
    let releaseNameValidation = null
    let releaseNameHelp = null
    if (releaseName !== releaseName.replace(/([^a-z0-9]+)/gi, '')) {
      releaseNameValidation = 'error'
      releaseNameHelp = 'Release name must be alphanumeric'
    } else if (!releaseName) {
      releaseNameValidation = 'warning'
      releaseNameHelp = 'Release name cannot be empty'
    }
    this.setState({ releaseName, releaseNameValidation, releaseNameHelp })
  }

  handleLinkChange = (e) => {
    this.setState({ releaseLink: e.target.value })
  }

  createRelease = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { releaseName, releaseNameValidation, releaseLink } = this.state
    if (releaseNameValidation) {
      return
    }
    const { params, spec } = this.props
    const release = {
      name: releaseName,
      link: releaseLink,
      waves: spec.waves
    }
    this.props.createRelease(params.env, release)
  }

  renderForm () {
    const { env } = this.props
    return (
      <form>
        <FormGroup>
          <ControlLabel>Deployed To</ControlLabel>
          <FormControl.Static>{env && env.deployedToEnv || ''}</FormControl.Static>
        </FormGroup>
        <FormGroup validationState={this.state.releaseNameValidation}>
          <ControlLabel>Name</ControlLabel>
          <FormControl
            type='text'
            value={this.state.releaseName}
            placeholder='Release name'
            onChange={this.handleNameChange}
          />
          <FormControl.Feedback />
          <HelpBlock>{this.state.releaseNameHelp}</HelpBlock>
        </FormGroup>
        <FormGroup>
          <ControlLabel>Link</ControlLabel>
          <FormControl
            type='text'
            value={this.state.releaseLink}
            placeholder='Release link'
            onChange={this.handleLinkChange}
          />
        </FormGroup>
      </form>
    )
  }

  renderWavePanels () {
    const { env, spec } = this.props
    if (spec) {
      return _.map(spec.waves, (wave, ix) => {
        return (
          <WavePanel
            key={ix}
            ix={ix}
            env={env.name}
            targets={wave.targets}
          />
        )
      })
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
          <li className='active'>New</li>
        </ol>
        <ButtonToolbar key='toolbar' className='pull-right'>
          <Button onClick={this.createRelease} bsStyle='primary' bsSize='small' disabled={!!this.state.releaseNameValidation}>Create</Button>
        </ButtonToolbar>
      </div>
    ]

    return (
      <div>
        {header}
        {this.renderForm()}
        <h5>Waves</h5>
        {this.renderWavePanels()}
      </div>
    )
  }

}

class WavePanel extends React.Component {
  static propTypes = {
    env: PropTypes.string,
    ix: PropTypes.number,
    targets: PropTypes.array
  }

  actionsTable () {
    const { targets } = this.props
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
    const { env, targets } = this.props
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
    const { env, targets } = this.props
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
