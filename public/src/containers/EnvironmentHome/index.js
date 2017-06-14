import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import { activateNav } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  return {
    env
  }
}

const dispatchProps = {
  activateNav
}

@connect(mapStateToProps, dispatchProps)
export default class EnvironmentHome extends React.Component {
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
    const { env } = this.props
    if (!env) {
      return null
    }
    const items = []

    items.push(
      <li key='releases'><Link to={`/${this.props.params.env}/releases`}>Releases</Link></li>)

    if (!_.isEmpty(env.deployments)) {
      items.push(
        <li key='deployments'><Link to={`/${this.props.params.env}/deployments`}>Deployments</Link></li>)
    }
    if (!_.isEmpty(env.jobs)) {
      items.push(
        <li key='jobs'><Link to={`/${this.props.params.env}/jobs`}>Jobs</Link></li>)
    }
    if (!_.isEmpty(env.configmaps)) {
      items.push(
        <li key='configmaps'><Link to={`/${this.props.params.env}/configmaps`}>Config Maps</Link></li>)
    }
    items.push(
      <li key='nodes'><Link to={`/${this.props.params.env}/nodes`}>Nodes</Link></li>)
    items.push(
      <li key='pods'><Link to={`/${this.props.params.env}/pods`}>Pods</Link></li>)
    return (
      <div>
        <div key='header' className='view-header'>
          <ol className='breadcrumb'>
            <li className='active'>{this.props.params.env}</li>
          </ol>
        </div>
        <ul key='list' className='nav nav-pills nav-stacked'>{items}</ul>
      </div>
    )
  }
}
