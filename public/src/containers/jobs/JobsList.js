import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Button, ButtonToolbar } from 'react-bootstrap'
import { Link } from 'react-router'
import _ from 'underscore'

import Table from '../../components/Table'
import { activateNav } from '../../actions/app'

function mapStateToProps (state, ownProps) {
  const env = _.findWhere(state.envs.toJS().envs, {name: ownProps.params.env})
  const jobRuns = state.jobRuns.lookUpObjects(ownProps.params.env)
  return {
    env,
    jobRuns
  }
}

@connect(mapStateToProps)
export default class JobsList extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.object,
    location: PropTypes.object,
    env: PropTypes.object,
    jobRuns: PropTypes.object
  }

  componentDidMount () {
    this.props.dispatch(activateNav('jobs'))
  }

  render () {
    const { params, env, jobRuns } = this.props

    const header = (
      <div className='view-header'>
        <ol className='breadcrumb'>
          <li><Link to={`/${params.env}`}>{params.env}</Link></li>
          <li className='active'>Jobs</li>
        </ol>
      </div>
    )

    if (env.approval) {
      header.push(
        <ButtonToolbar key='toolbar' pullRight>
          <Button onClick={this.release} bsStyle='success' bsSize='small'>Release</Button>
        </ButtonToolbar>)
    }

    const columns = [
      {title: 'Name', key: 'name'},
      {title: 'Tag', key: 'tag', style: {width: '180px'}},
      {title: 'Last Run', key: 'lastRun', style: {width: '200px', textAlign: 'right'}}
    ]

    var rows = _.map(env.jobs, (jobName) => {
      const runs = _.sortBy(
        _.filter(jobRuns, x => x.hasLabel('job', jobName)),
        x => -x.creationTimestamp
      )
      const jobRun = runs[0]
      return {
        name: (<Link to={`/${env.name}/jobs/${jobName}`}>{jobName}</Link>),
        tag: jobRun && jobRun.imageTag,
        lastRun: jobRun && jobRun.runAt
      }
    })

    return (
      <div>
        {header}
        <Table columns={columns} rows={rows} />
      </div>
    )
  }

}
