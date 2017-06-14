import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Button, ButtonToolbar, Label } from 'react-bootstrap'
import { browserHistory, Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'
import { createReleaseFromLatest, deployRelease, deleteRelease } from '../../actions/releases'

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  const releases = state.releases.lookUpObjects(ownProps.params.env)
  return {
    env,
    releases
  }
}

const dispatchProps = {
  activateNav,
  createReleaseFromLatest
}

@connect(mapStateToProps, dispatchProps)
export default class ReleasesList extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    env: PropTypes.object,
    releases: PropTypes.object,
    activateNav: PropTypes.func.isRequired,
    createReleaseFromLatest: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateNav('releases')
  }

  goToCreate = () => {
    const { params } = this.props
    browserHistory.push(`/${params.env}/releases/create`)
  }

  createLatest = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { params, createReleaseFromLatest } = this.props
    createReleaseFromLatest(params.env)
  }

  renderHeader () {
    const { params, env } = this.props
    const style = {marginRight: '10px'}
    const buttons = []
    buttons.push(
      <Button key='latest' onClick={this.createLatest} style={style} bsStyle='primary' bsSize='small'>Create from Latest Versions</Button>
    )
    if (env.deployedToEnv) {
      buttons.push(
        <Button key='current' onClick={this.goToCreate} style={style} bsStyle='success' bsSize='small'>Create from Current Versions</Button>
      )
    }
    return (
      <div key='header' className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Releases</li>
        </ol>
        <ButtonToolbar key='toolbar' className='pull-right'>
          {buttons}
        </ButtonToolbar>
      </div>
    )
  }

  render () {
    const { env, releases } = this.props
    const columns = [
      {title: 'Name', key: 'name'},
      {title: 'Link', key: 'link', style: {width: '200px'}},
      {title: 'Approved By', key: 'createdBy', style: {width: '150px'}},
      {title: 'Created At', key: 'createdAt', style: {width: '150px'}},
      {title: 'Deployments', key: 'deployments', style: {width: '150px'}},
      {title: 'Actions', key: 'actions', style: {width: '150px', textAlign: 'right'}}
    ]

    const rows = _.map(releases, function (release) {
      return {
        component: (
          <Row key={release.name}
            env={env}
            release={release}
          />
        ),
        key: -(new Date(release.createdAt))
      }
    })
    const sortedRows = _.sortBy(rows, 'key')

    return (
      <div>
        {this.renderHeader()}
        <Table columns={columns} rows={sortedRows} />
      </div>
    )
  }
}

const rowDispatchProps = {
  deployRelease,
  deleteRelease
}

@connect(null, rowDispatchProps)
class Row extends React.Component {
  static propTypes = {
    deployRelease: PropTypes.func.isRequired,
    deleteRelease: PropTypes.func.isRequired,
    env: PropTypes.object.isRequired,
    release: PropTypes.object.isRequired
  }

  get nameLink () {
    const { env, release } = this.props
    return (
      <Link to={`/${env.name}/releases/${release.name}`}>{release.name}</Link>
    )
  }

  get link () {
    const { release } = this.props
    if (!release.link) {
      return null
    }
    return (
      <a href={release.link} target='_blank'>{release.link}</a>
    )
  }

  deployRelease = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { deployRelease, env, release } = this.props
    deployRelease(env.name, release.name)
  }

  deleteRelease = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { deleteRelease, env, release } = this.props
    deleteRelease(env.name, release.name)
  }

  get actions () {
    const { env } = this.props
    const style = {
      marginLeft: '10px'
    }
    const actions = []
    if (env.deployedToEnv) {
      actions.push(
        <Button onClick={this.deployRelease} style={style} bsStyle='primary' bsSize='xs'>Deploy</Button>
      )
      actions.push(
        <Button onClick={this.deleteRelease} style={style} bsStyle='danger' bsSize='xs'>Delete</Button>
      )
    } else if (env.approvedFromEnv) {
      actions.push(
        <Button onClick={this.deployRelease} style={style} bsStyle='primary' bsSize='xs'>Deploy</Button>
      )
    }
    return actions
  }

  render () {
    const { env, release } = this.props
    const releasedAt = _.map(release.envRollouts(env.name), (rollout) => {
      let bsStyle = 'default'
      switch (rollout.status) {
        case 'deployed':
          bsStyle = 'success'
          break
        case 'deploying':
          bsStyle = 'warning'
          break
        case 'failed':
          bsStyle = 'danger'
          break
      }
      return (
        <div key={rollout.id}>
          <Label bsStyle={bsStyle}>{rollout.id} - {rollout.rolloutAtHumanize}</Label>
        </div>
      )
    })
    return (
      <tr>
        <td>{this.nameLink}</td>
        <td>{this.link}</td>
        <td>{release.createdBy || '-'}</td>
        <td>{release.createdAtHumanize || '-'}</td>
        <td>{releasedAt}</td>
        <td style={{textAlign: 'right'}}>{this.actions}</td>
      </tr>
    )
  }

}
