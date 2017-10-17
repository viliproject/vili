import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'

import Loading from '../../components/Loading'
import { activateDeploymentTab } from '../../actions/app'
import { getDeploymentSpec } from '../../actions/deployments'

function mapStateToProps (state, ownProps) {
  const { env, deployment: deploymentName } = ownProps.params
  const deployment = state.deployments.lookUpData(env, deploymentName)
  return {
    deployment
  }
}

const dispatchProps = {
  activateDeploymentTab,
  getDeploymentSpec
}

export class DeploymentSpec extends React.Component {
  static propTypes = {
    params: PropTypes.object,
    location: PropTypes.object,
    deployment: PropTypes.object,
    activateDeploymentTab: PropTypes.func,
    getDeploymentSpec: PropTypes.func
  }

  componentDidMount () {
    this.props.activateDeploymentTab('spec')
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params, getDeploymentSpec } = this.props
    getDeploymentSpec(params.env, params.deployment)
  }

  render () {
    const { deployment } = this.props
    if (!deployment || !deployment.get('spec')) {
      return (<Loading />)
    }
    return (
      <div className='col-md-8'>
        <div id='source-yaml'>
          <pre><code>
            {deployment.get('spec')}
          </code></pre>
        </div>
      </div>
    )
  }

}

export default connect(mapStateToProps, dispatchProps)(DeploymentSpec)
