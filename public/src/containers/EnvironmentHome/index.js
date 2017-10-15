import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import { activateNav } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const { env: envName } = ownProps.params
  const env = state.envs.getIn(['envs', envName])
  return {
    env
  }
}

const dispatchProps = {
  activateNav
}

export class EnvironmentHome extends React.Component {
  static propTypes = {
    env: PropTypes.object,
    activateNav: PropTypes.func.isRequired,
    params: PropTypes.object,
    location: PropTypes.object
  }

  componentDidMount () {
    this.props.activateNav('')
  }

  render () {
    const { env, params } = this.props
    if (!env) {
      return null
    }
    const items = []

    items.push(
      <li key='releases'><Link to={`/${params.env}/releases`}>Releases</Link></li>)

    if (env.deployments && !env.deployments.isEmpty()) {
      items.push(
        <li key='deployments'><Link to={`/${params.env}/deployments`}>Deployments</Link></li>)
    }
    if (env.jobs && !env.jobs.isEmpty()) {
      items.push(
        <li key='jobs'><Link to={`/${params.env}/jobs`}>Jobs</Link></li>)
    }
    if (env.configmaps && !env.configmaps.isEmpty()) {
      items.push(
        <li key='configmaps'><Link to={`/${params.env}/configmaps`}>Config Maps</Link></li>)
    }
    items.push(
      <li key='nodes'><Link to={`/${params.env}/nodes`}>Nodes</Link></li>)
    items.push(
      <li key='pods'><Link to={`/${params.env}/pods`}>Pods</Link></li>)
    return (
      <div>
        <div key='header' className='view-header'>
          <ol className='breadcrumb'>
            <li className='active'>{params.env}</li>
          </ol>
        </div>
        <ul key='list' className='nav nav-pills nav-stacked'>{items}</ul>
      </div>
    )
  }
}

export default connect(mapStateToProps, dispatchProps)(EnvironmentHome)
