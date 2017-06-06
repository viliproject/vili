import React, { PropTypes } from 'react'
import { connect } from 'react-redux'
import { Button, Label } from 'react-bootstrap'
import _ from 'underscore'

import { deployTag } from '../../actions/deployments'

const dispatchProps = {
  deployTag
}

@connect(null, dispatchProps)
export default class DeploymentRow extends React.Component {
  static propTypes = {
    deployTag: PropTypes.func.isRequired,
    env: PropTypes.string,
    deployment: PropTypes.string,
    currentRevision: PropTypes.number,
    tag: PropTypes.string,
    branch: PropTypes.string,
    revision: PropTypes.string,
    buildTime: PropTypes.string,
    replicaSets: PropTypes.object
  }

  deployTag = (event) => {
    event.target.setAttribute('disabled', 'disabled')
    const { deployTag, env, deployment, tag, branch } = this.props
    deployTag(env, deployment, tag, branch)
  }

  render () {
    const { currentRevision, tag, branch, revision, buildTime, replicaSets } = this.props
    var className = ''
    const deployedAt = _.map(replicaSets, (replicaSet) => {
      var bsStyle = 'default'
      if (currentRevision === replicaSet.revision) {
        className = 'success'
        bsStyle = 'success'
      } else if (replicaSet.status.replicas > 0) {
        className = 'warning'
        bsStyle = 'warning'
      }
      return (
        <div key={replicaSet.metadata.name}>
          <Label bsStyle={bsStyle}>{replicaSet.revision} - {replicaSet.deployedAt}</Label>
        </div>
      )
    })
    return (
      <tr className={className}>
        <td>{tag}</td>
        <td>{branch}</td>
        <td>{revision || 'unknown'}</td>
        <td>{buildTime}</td>
        <td style={{textAlign: 'right'}}>{deployedAt}</td>
        <td style={{textAlign: 'right'}}>
          <Button onClick={this.deployTag} bsStyle='primary' bsSize='xs'>Deploy</Button>
        </td>
      </tr>
    )
  }

}
