import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { Navbar, Nav, NavDropdown, MenuItem } from 'react-bootstrap'

import LinkMenuItem from '../../components/LinkMenuItem'
import EnvCreateModal from '../EnvCreateModal'
import { showCreateEnvModal, deleteEnvironment } from '../../actions/envs'

function mapStateToProps (state, ownProps) {
  const envs = state.envs.get('envs')
  const env = envs.get(ownProps.envName)
  return {
    app: state.app,
    user: state.user,
    envs,
    env
  }
}

const dispatchProps = {
  showCreateEnvModal,
  deleteEnvironment
}

@connect(mapStateToProps, dispatchProps)
export default class TopNav extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    location: PropTypes.object,
    user: PropTypes.object,
    envs: PropTypes.object,
    env: PropTypes.object,
    showCreateEnvModal: PropTypes.func.isRequired,
    deleteEnvironment: PropTypes.func.isRequired
  }

  showCreateEnvModal = () => {
    this.props.showCreateEnvModal()
  }

  renderLoggedInNav () {
    const { user, envs, env } = this.props

    // user
    const userText = user.firstName + ' ' + user.lastName + ' (' + user.username + ')'

    // environments
    const envElements = []
    envs.map((e) => {
      var onRemove = null
      if (env && e.name !== env.name && !e.protected) {
        onRemove = () => {
          this.props.deleteEnvironment(e.name)
        }
      }
      envElements.push(
        <LinkMenuItem
          key={e.name}
          to={`/${e.name}`}
          active={env && env.name === e.name}
          onRemove={onRemove}
        >
          {e.name}
        </LinkMenuItem>
      )
    })

    return (
      <Navbar className={env && env.approvedFromEnv ? 'prod' : ''}
        fixedTop fluid
      >
        <div className='navbar-header pull-left'>
          <Link className='navbar-brand' to='/'>Vili</Link>
        </div>
        <Nav key='user' className='user' pullRight>
          <NavDropdown id='user-dropdown' title={userText}>
            <MenuItem title='Logout' href='/logout'>Logout</MenuItem>
          </NavDropdown>
        </Nav>
        <Nav key='env' className='environment' pullRight>
          <NavDropdown id='env-dropdown'
            title={(env && env.name) || <span className='text-danger'>Select Environment</span>}
          >
            {envElements}
            <MenuItem divider />
            <MenuItem onSelect={this.showCreateEnvModal}>Create Environment</MenuItem>
          </NavDropdown>
        </Nav>
        <EnvCreateModal />
      </Navbar>
    )
  }

  renderLoggedOutNav () {
    return (
      <Navbar fixedTop fluid>
        <div className='navbar-header pull-left'>
          <Link className='navbar-brand' to='/'>Vili</Link>
        </div>
        <Nav className='user' pullRight>
          <MenuItem title='Login' href='/login'>Login</MenuItem>
        </Nav>
      </Navbar>
    )
  }

  render () {
    if (this.props.user) {
      return this.renderLoggedInNav()
    }
    return this.renderLoggedOutNav()
  }

}
