import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"

import Table from "../../components/Table"
import { activateNav } from "../../actions/app"
import { makeLookUpObjects } from "../../selectors"

import ConfigMapsListRow from "./ConfigMapsListRow"

function makeMapStateToProps() {
  const lookUpObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { envName } = ownProps
    const env = state.envs.getIn(["envs", envName])
    const configmaps = lookUpObjects(state.configmaps, envName)
    return {
      env,
      configmaps,
    }
  }
}

const dispatchProps = {
  activateNav,
}

export class ConfigMapsList extends React.Component {
  componentDidMount() {
    this.props.activateNav("configmaps")
  }

  render() {
    const { envName, env, configmaps } = this.props
    const header = (
      <div className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li className="active">Config Maps</li>
        </ol>
      </div>
    )

    const columns = [
      { title: "Name", key: "name" },
      { title: "Key Count", key: "key-count" },
      { title: "Created", key: "created" },
    ]

    const rows = []
    env.configmaps.forEach(configmapName => {
      const configmap = configmaps.find(
        d => d.getIn(["metadata", "name"]) === configmapName
      )
      rows.push({
        component: (
          <ConfigMapsListRow
            key={configmapName}
            envName={envName}
            name={configmapName}
            configmap={configmap}
          />
        ),
        key: configmapName,
      })
    })

    return (
      <div>
        {header}
        <Table columns={columns} rows={rows} />
      </div>
    )
  }
}

ConfigMapsList.propTypes = {
  envName: PropTypes.string,
  env: PropTypes.object,
  configmaps: PropTypes.object,
  activateNav: PropTypes.func,
}

export default connect(makeMapStateToProps, dispatchProps)(ConfigMapsList)
