import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button, Label } from "react-bootstrap"
import { Link } from "react-router-dom"

import { deployRelease, deleteRelease } from "../../actions/releases"

const rowDispatchProps = {
  deployRelease,
  deleteRelease,
}

export class ReleasesListRow extends React.Component {
  get nameLink() {
    const { env, release } = this.props
    return (
      <Link to={`/${env.name}/releases/${release.name}`}>{release.name}</Link>
    )
  }

  get link() {
    const { release } = this.props
    if (!release.link) {
      return null
    }
    return (
      <a href={release.link} target="_blank">
        {release.link}
      </a>
    )
  }

  deployRelease = event => {
    event.target.setAttribute("disabled", "disabled")
    const { deployRelease, env, release } = this.props
    deployRelease(env.name, release.name)
  }

  deleteRelease = event => {
    event.target.setAttribute("disabled", "disabled")
    const { deleteRelease, env, release } = this.props
    deleteRelease(env.name, release.name)
  }

  get actions() {
    const { env } = this.props
    const style = {
      marginLeft: "10px",
    }
    const actions = []
    if (env.deployedToEnv) {
      actions.push(
        <Button
          key="deploy"
          onClick={this.deployRelease}
          style={style}
          bsStyle="primary"
          bsSize="xs"
        >
          Deploy
        </Button>
      )
      actions.push(
        <Button
          key="delete"
          onClick={this.deleteRelease}
          style={style}
          bsStyle="danger"
          bsSize="xs"
        >
          Delete
        </Button>
      )
    } else if (env.approvedFromEnv) {
      actions.push(
        <Button
          key="deploy"
          onClick={this.deployRelease}
          style={style}
          bsStyle="primary"
          bsSize="xs"
        >
          Deploy
        </Button>
      )
    }
    return actions
  }

  render() {
    const { env, release } = this.props
    const releasedAt = []
    release
      .envRollouts(env.name)
      .sortBy(x => -x.rolloutAtDate)
      .forEach(rollout => {
        let bsStyle = "default"
        switch (rollout.status) {
          case "deployed":
            bsStyle = "success"
            break
          case "deploying":
            bsStyle = "warning"
            break
          case "failed":
            bsStyle = "danger"
            break
        }
        releasedAt.push(
          <div key={rollout.id}>
            <Label bsStyle={bsStyle}>
              {rollout.id} - {rollout.rolloutAtHumanize}
            </Label>
          </div>
        )
      })
    return (
      <tr>
        <td>{this.nameLink}</td>
        <td>{this.link}</td>
        <td>{release.createdBy || "-"}</td>
        <td>{release.createdAtHumanize || "-"}</td>
        <td>{releasedAt}</td>
        <td style={{ textAlign: "right" }}>{this.actions}</td>
      </tr>
    )
  }
}

ReleasesListRow.propTypes = {
  deployRelease: PropTypes.func.isRequired,
  deleteRelease: PropTypes.func.isRequired,
  env: PropTypes.object.isRequired,
  release: PropTypes.object.isRequired,
}

export default connect(null, rowDispatchProps)(ReleasesListRow)
