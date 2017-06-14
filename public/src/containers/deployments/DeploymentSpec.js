import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'

import Loading from '../../components/Loading'
import { activateDeploymentTab } from '../../actions/app'
import { getDeploymentSpec } from '../../actions/deployments'

function mapStateToProps (state, ownProps) {
  const deployment = state.deployments.lookUpData(ownProps.params.env, ownProps.params.deployment)
  return {
    deployment
  }
}

@connect(mapStateToProps)
export default class DeploymentSpec extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object, // react router provides this
    location: PropTypes.object, // react router provides this
    deployment: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateDeploymentTab('spec'))
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params } = this.props
    this.props.dispatch(getDeploymentSpec(params.env, params.deployment))
  }

  render () {
    const { deployment } = this.props
    if (!deployment || !deployment.spec) {
      return (<Loading />)
    }
    return (
      <div className='col-md-8'>
        <div id='source-yaml'>
          <pre><code>
            {deployment.spec}
          </code></pre>
        </div>
      </div>
    )
  }

}
