import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'

import Loading from '../../components/Loading'
import { activateJobTab } from '../../actions/app'
import { getJobSpec } from '../../actions/jobs'

function mapStateToProps (state, ownProps) {
  const { env, job: jobName } = ownProps.params
  const job = state.jobs.lookUpData(env, jobName)
  return {
    job
  }
}

const dispatchProps = {
  activateJobTab,
  getJobSpec
}

export class JobSpec extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    job: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateJobTab('spec'))
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params } = this.props
    this.props.dispatch(getJobSpec(params.env, params.job))
  }

  render () {
    const { job } = this.props
    if (!job || !job.spec) {
      return (<Loading />)
    }
    return (
      <div className='col-md-8'>
        <div id='source-yaml'>
          <pre><code>
            {job.spec}
          </code></pre>
        </div>
      </div>
    )
  }
}

export default connect(mapStateToProps, dispatchProps)(JobSpec)
