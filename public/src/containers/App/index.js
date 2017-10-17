import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'

import TopNav from '../../containers/TopNav'
import SideNav from '../../components/SideNav'

function mapStateToProps (state, ownProps) {
  const { env: envName } = ownProps.params
  const env = state.envs.getIn(['envs', envName])
  return {
    app: state.app,
    env
  }
}

export class App extends React.Component {
  static propTypes = {
    children: PropTypes.node,
    location: PropTypes.object,
    params: PropTypes.object,
    app: PropTypes.object,
    env: PropTypes.object
  }

  render () {
    const { app, env, children, params } = this.props
    return (
      <div className='top-nav container-fluid full-height'>
        <TopNav location={location} envName={params.env} />
        <div className='page-wrapper'>
          <div className='sidebar'>
            <SideNav env={env} nav={app.get('nav')} />
          </div>
          <div className='content-wrapper'>{children}</div>
        </div>
      </div>
    )
  }
}

export default connect(mapStateToProps)(App)
