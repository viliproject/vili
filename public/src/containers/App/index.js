import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import _ from 'underscore'

import TopNav from '../../containers/TopNav'
import SideNav from '../../components/SideNav'

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  return {
    app: state.app.toJS(),
    env
  }
}

@connect(mapStateToProps)
export default class App extends React.Component {
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
            <SideNav env={env} nav={app.nav} />
          </div>
          <div className='content-wrapper'>{children}</div>
        </div>
      </div>
    )
  }
}
