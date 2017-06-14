import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Button, Label } from 'react-bootstrap'
import _ from 'underscore'

import { runTag } from '../../actions/jobs'

@connect()
export default class JobRow extends React.Component {
  static propTypes = {
    dispatch: PropTypes.func,
    env: PropTypes.string,
    job: PropTypes.string,
    tag: PropTypes.string,
    branch: PropTypes.string,
    revision: PropTypes.string,
    buildTime: PropTypes.string,
    jobRuns: PropTypes.array
  }

  runTag = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { dispatch, env, job, tag, branch } = this.props
    dispatch(runTag(env, job, tag, branch))
  }

  render () {
    const { tag, branch, revision, buildTime, jobRuns } = this.props
    const runTimes = _.map(jobRuns, function (jobRun) {
      var bsStyle = 'default'
      _.each(jobRun.status.conditions, (condition) => {
        switch (condition.type) {
          case 'Complete':
            bsStyle = 'success'
            break
          case 'Failed':
            bsStyle = 'danger'
            break
        }
      })
      return (
        <div key={jobRun.metadata.name}>
          <Label bsStyle={bsStyle}>{jobRun.runAt}</Label>
        </div>
      )
    })
    return (
      <tr>
        <td>{tag}</td>
        <td>{branch}</td>
        <td>{revision || 'unknown'}</td>
        <td>{buildTime}</td>
        <td style={{textAlign: 'right'}}>{runTimes}</td>
        <td style={{textAlign: 'right'}}>
          <Button onClick={this.runTag} bsStyle='primary' bsSize='xs'>Run</Button>
        </td>
      </tr>
    )
  }

}
