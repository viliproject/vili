import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import _ from 'underscore'

import { activateNav } from '../../actions/app'

const tabs = {
  'home': 'Home',
  'rollouts': 'Rollouts',
  'spec': 'Spec',
  'service': 'Service'
}

function mapStateToProps (state) {
  return {
    app: state.app
  }
}

const dispatchProps = {
  activateNav
}

export class DeploymentBase extends React.Component {
  static propTypes = {
    children: PropTypes.node,
    params: PropTypes.object,
    location: PropTypes.object,
    app: PropTypes.object,
    activateNav: PropTypes.func.isRequired
  }

  componentDidMount () {
    const { params, activateNav } = this.props
    activateNav('deployments', params.deployment)
  }

  componentDidUpdate (prevProps) {
    const { params, activateNav } = this.props
    if (params.deployment !== prevProps.params.deployment) {
      activateNav('deployments', params.deployment)
    }
  }

  render () {
    const { params, app, children } = this.props
    const tabElements = _.map(tabs, (name, key) => {
      let className = ''
      if (app.get('deploymentTab') === key) {
        className = 'active'
      }
      let link = `/${params.env}/deployments/${params.deployment}`
      if (key !== 'home') {
        link += `/${key}`
      }
      return (
        <li key={key} role='presentation' className={className}>
          <Link to={link}>{name}</Link>
        </li>
      )
    })
    return (
      <div>
        <div key='view-header' className='view-header'>
          <ol className='breadcrumb'>
            <li><Link to={`/${params.env}`}>{params.env}</Link></li>
            <li><Link to={`/${params.env}/deployments`}>Deployments</Link></li>
            <li className='active'>{params.deployment}</li>
          </ol>
          <ul className='nav nav-pills pull-right'>
            {tabElements}
          </ul>
        </div>
        {children}
      </div>
    )
  }
}

export default connect(mapStateToProps, dispatchProps)(DeploymentBase)
