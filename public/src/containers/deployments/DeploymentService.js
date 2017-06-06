import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import Loading from '../../components/Loading'
import { activateDeploymentTab } from '../../actions/app'
import { getDeploymentService } from '../../actions/deployments'

function mapStateToProps (state, ownProps) {
  const deployment = state.deployments.lookUpData(ownProps.params.env, ownProps.params.deployment)
  return {
    deployment
  }
}

@connect(mapStateToProps)
export default class DeploymentService extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object, // react router provides this
    location: PropTypes.object, // react router provides this
    deployment: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateDeploymentTab('service'))
    this.subData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.subData()
    }
  }

  subData = () => {
    const { params } = this.props
    this.props.dispatch(getDeploymentService(params.env, params.deployment))
  }

  clickCreateService = (event) => {
    var self = this
    event.currentTarget.setAttribute('disabled', 'disabled')
    viliApi.services.create(this.props.params.env, this.props.params.app).then(function () {
      self.loadData()
    })
  }

  render () {
    const { deployment } = this.props
    if (!deployment) {
      return (<Loading />)
    }
    if (!deployment.service) {
      return (
        <div id='service'>
          <div className='alert alert-warning' role='alert'>No Service Defined</div>
          <div><button className='btn btn-success' onClick={this.clickCreateService}>Create Service</button></div>
        </div>
      )
    }
    return (
      <div id='service'>
        IP: {deployment.service.spec.clusterIP}
      </div>
    )
  }
}
