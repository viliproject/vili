import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Button, ButtonToolbar } from "react-bootstrap"
import { Link } from "react-router-dom"
import _ from "underscore"

import { makeLookUpObjects } from "../../selectors"
import Table from "../../components/Table"
import { activateNav } from "../../actions/app"
import { createReleaseFromLatest } from "../../actions/releases"
import history from "../../lib/history"

import ReleasesListRow from "./ReleasesListRow"

function makeMapStateToProps() {
  const lookUpObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { envName } = ownProps
    const env = state.envs.getIn(["envs", envName])
    const releases = lookUpObjects(state.releases, envName)
    return {
      env,
      releases,
    }
  }
}

const dispatchProps = {
  activateNav,
  createReleaseFromLatest,
}

export class ReleasesList extends React.Component {
  componentDidMount() {
    this.props.activateNav("releases")
  }

  goToCreate = () => {
    const { envName } = this.props
    history.push(`/${envName}/releases/create`)
  }

  createLatest = event => {
    event.target.setAttribute("disabled", "disabled")
    const { envName, createReleaseFromLatest } = this.props
    createReleaseFromLatest(envName)
  }

  renderHeader() {
    const { envName, env } = this.props
    const style = { marginRight: "10px" }
    const buttons = []
    buttons.push(
      <Button
        key="latest"
        onClick={this.createLatest}
        style={style}
        bsStyle="primary"
        bsSize="small"
      >
        Create from Latest Versions
      </Button>
    )
    if (env.deployedToEnv) {
      buttons.push(
        <Button
          key="current"
          onClick={this.goToCreate}
          style={style}
          bsStyle="success"
          bsSize="small"
        >
          Create from Current Versions
        </Button>
      )
    }
    return (
      <div key="header" className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li className="active">Releases</li>
        </ol>
        <ButtonToolbar key="toolbar" className="pull-right">
          {buttons}
        </ButtonToolbar>
      </div>
    )
  }

  render() {
    const { env, releases } = this.props
    const columns = [
      { title: "Name", key: "name" },
      { title: "Link", key: "link", style: { width: "200px" } },
      { title: "Approved By", key: "createdBy", style: { width: "150px" } },
      { title: "Created At", key: "createdAt", style: { width: "150px" } },
      { title: "Deployments", key: "deployments", style: { width: "150px" } },
      {
        title: "Actions",
        key: "actions",
        style: { width: "150px", textAlign: "right" },
      },
    ]

    const rows = []
    releases.map(release => {
      rows.push({
        component: (
          <ReleasesListRow key={release.name} env={env} release={release} />
        ),
        key: -new Date(release.createdAt),
      })
    })
    const sortedRows = _.sortBy(rows, "key")

    return (
      <div>
        {this.renderHeader()}
        <Table columns={columns} rows={sortedRows} />
      </div>
    )
  }
}

ReleasesList.propTypes = {
  envName: PropTypes.string,
  env: PropTypes.object,
  releases: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  createReleaseFromLatest: PropTypes.func.isRequired,
}

export default connect(makeMapStateToProps, dispatchProps)(ReleasesList)
