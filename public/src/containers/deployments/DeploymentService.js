import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'

import Loading from '../../components/Loading'
import { activateDeploymentTab } from '../../actions/app'
import { getDeploymentService } from '../../actions/deployments'
import { createService } from '../../actions/services'

function mapStateToProps (state, ownProps) {
  const { env, deployment: deploymentName } = ownProps.params
  const deployment = state.deployments.lookUpData(env, deploymentName)
  return {
    deployment
  }
}

const dispatchProps = {
  activateDeploymentTab,
  getDeploymentService,
  createService
}

export class DeploymentService extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    deployment: PropTypes.object,
    activateDeploymentTab: PropTypes.func.isRequired,
    getDeploymentService: PropTypes.func.isRequired,
    createService: PropTypes.func.isRequired
  }

  componentDidMount () {
    this.props.activateDeploymentTab('service')
    this.subData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.subData()
    }
  }

  subData = () => {
    const { params } = this.props
    this.props.getDeploymentService(params.env, params.deployment)
  }

  clickCreateService = (event) => {
    const { params } = this.props
    event.currentTarget.setAttribute('disabled', 'disabled')
    this.props.createService(params.env, params.app)
  }

  render () {
    const { deployment } = this.props
    if (!deployment) {
      return (<Loading />)
    }
    if (!deployment.get('service')) {
      return (
        <div id='service'>
          <div className='alert alert-warning' role='alert'>No Service Defined</div>
          <div><button className='btn btn-success' onClick={this.clickCreateService}>Create Service</button></div>
        </div>
      )
    }
    return (
      <div id='service'>
        IP: {deployment.getIn(['service', 'spec', 'clusterIP'])}
      </div>
    )
  }
}

export default connect(mapStateToProps, dispatchProps)(DeploymentService)
