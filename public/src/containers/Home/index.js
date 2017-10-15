import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'

import { activateNav } from '../../actions/app'

function mapStateToProps (state) {
  return {
    envs: state.envs.get('envs')
  }
}

const dispatchProps = {
  activateNav
}

export class Home extends React.Component {
  static propTypes = {
    envs: PropTypes.object,
    activateNav: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateNav('home')
  }

  renderEnvs () {
    const { envs } = this.props
    const links = []
    envs.forEach((env) => {
      links.push(<li key={env.name}><Link to={`/${env.name}`}>{env.name}</Link></li>)
    })
    return (<div>
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li className='active'>Select Environment</li>
        </ol>
      </div>
      <ul className='nav nav-pills nav-stacked'>{links}</ul>
    </div>)
  }

  renderLoggedOut () {
    return (
      <div className='jumbotron'>
        <h1>Welcome to Vili</h1>
        <p>Please log in to view your applications.</p>
        <p><a className='btn btn-primary btn-lg' href='/login' role='button'>Login</a></p>
      </div>
    )
  }

  render () {
    const { envs } = this.props
    if (!envs || envs.isEmpty()) {
      return this.renderLoggedOut()
    }
    return this.renderEnvs()
  }
}

export default connect(mapStateToProps, dispatchProps)(Home)
