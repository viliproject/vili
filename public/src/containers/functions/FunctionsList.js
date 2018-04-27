import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"

import Table from "../../components/Table"
import { activateNav } from "../../actions/app"
import { makeLookUpObjects } from "../../selectors"

function makeMapStateToProps() {
  const lookUpFunctionObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { envName } = ownProps
    const env = state.envs.getIn(["envs", envName])
    const functions = lookUpFunctionObjects(state.functions, env.name)
    return {
      env,
      functions,
    }
  }
}

const dispatchProps = {
  activateNav,
}

export class FunctionsList extends React.Component {
  componentDidMount() {
    this.props.activateNav("functions")
  }

  render() {
    const { envName, env, functions } = this.props

    const header = (
      <div className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li className="active">Functions</li>
        </ol>
      </div>
    )

    const columns = [
      { title: "Name", key: "name" },
      { title: "Tag", key: "tag", style: { width: "250px" } },
      {
        title: "Deployed",
        key: "deployedAt",
        style: { width: "200px", textAlign: "right" },
      },
    ]

    const rows = []
    env.functions.forEach(functionName => {
      const func = functions.find(f => f.name === functionName)
      rows.push({
        name: (
          <Link to={`/${env.name}/functions/${functionName}`}>
            {functionName}
          </Link>
        ),
        tag: func && func.activeVersion && func.activeVersion.tag,
        deployedAt: func && func.activeVersion && func.activeVersion.deployedAt,
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

FunctionsList.propTypes = {
  envName: PropTypes.string.isRequired,
  env: PropTypes.object.isRequired,
  functions: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
}

export default connect(makeMapStateToProps, dispatchProps)(FunctionsList)
