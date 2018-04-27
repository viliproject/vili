import React from "react"
import PropTypes from "prop-types"
import { Panel } from "react-bootstrap"
import { Link } from "react-router-dom"
import _ from "underscore"

import Table from "../../components/Table"

const tableColumns = {
  waveActions: [
    { title: "Action", key: "name" },
    {
      title: "Branch",
      key: "branch",
      style: { width: "200px", textAlign: "right" },
    },
  ],
  waveJobs: [
    { title: "Job", key: "name" },
    {
      title: "Branch",
      key: "branch",
      style: { width: "200px", textAlign: "right" },
    },
    { title: "Tag", key: "tag", style: { width: "200px", textAlign: "right" } },
  ],
  waveApps: [
    { title: "App", key: "name" },
    {
      title: "Branch",
      key: "branch",
      style: { width: "200px", textAlign: "right" },
    },
    { title: "Tag", key: "tag", style: { width: "200px", textAlign: "right" } },
  ],
}

export class ReleaseWavePanel extends React.Component {
  actionsTable() {
    const { targets } = this.props.wave
    const rows = _.map(
      _.filter(targets, target => target.type === "action"),
      target => {
        return {
          name: target.name,
          branch: target.branch,
        }
      }
    )
    if (rows.length > 0) {
      return (
        <Table
          columns={tableColumns.waveActions}
          rows={rows}
          fill
          hover={false}
        />
      )
    }
    return null
  }

  jobsTable() {
    const { env } = this.props
    const { targets } = this.props.wave
    const rows = _.map(
      _.filter(targets, target => target.type === "job"),
      target => {
        return {
          name: <Link to={`/${env}/jobs/${target.name}`}>{target.name}</Link>,
          branch: target.branch,
          tag: target.tag,
          runAt: target.runAt,
        }
      }
    )
    if (rows.length > 0) {
      return (
        <Table columns={tableColumns.waveJobs} rows={rows} fill hover={false} />
      )
    }
    return null
  }

  appsTable() {
    const { env } = this.props
    const { targets } = this.props.wave
    const rows = _.map(
      _.filter(targets, target => target.type === "app"),
      target => {
        return {
          name: (
            <Link to={`/${env}/deployments/${target.name}`}>{target.name}</Link>
          ),
          branch: target.branch,
          tag: target.tag,
          deployedAt: target.deployedAt,
        }
      }
    )
    if (rows.length > 0) {
      return (
        <Table columns={tableColumns.waveApps} rows={rows} fill hover={false} />
      )
    }
    return null
  }

  render() {
    const { ix } = this.props
    return (
      <Panel>
        <Panel.Heading>{`Wave ${ix + 1}`}</Panel.Heading>
        {this.actionsTable()}
        {this.jobsTable()}
        {this.appsTable()}
      </Panel>
    )
  }
}

ReleaseWavePanel.propTypes = {
  env: PropTypes.string,
  ix: PropTypes.number,
  wave: PropTypes.object,
}

export default ReleaseWavePanel
