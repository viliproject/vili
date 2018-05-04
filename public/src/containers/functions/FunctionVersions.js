import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"

import Table from "../../components/Table"
import Loading from "../../components/Loading"
import { activateFunctionTab } from "../../actions/app"

import FunctionVersionRow from "./FunctionVersionRow"

function makeMapStateToProps() {
  return (state, ownProps) => {
    const { envName, functionName } = ownProps
    const func = state.functions.lookUpData(envName, functionName)
    return {
      func,
    }
  }
}

const dispatchProps = {
  activateFunctionTab,
}

export class FunctionVersions extends React.Component {
  componentDidMount() {
    this.props.activateFunctionTab("versions")
  }

  render() {
    const { envName, functionName, func } = this.props
    if (!func || !func.get("object")) {
      return <Loading />
    }

    const columns = [
      { title: "Revision", key: "revision", style: { width: "90px" } },
      { title: "Tag", key: "tag", style: { width: "180px" } },
      { title: "Branch", key: "branch" },
      { title: "Deployed At", key: "time", style: { textAlign: "right" } },
      {
        title: "Deployed By",
        key: "deployedBy",
        style: { textAlign: "right" },
      },
      { title: "Actions", key: "actions", style: { textAlign: "right" } },
    ]

    const rows = []
    const activeVersion = func.get("object").activeVersion
    func.get("object").versions.forEach(version => {
      rows.push({
        component: (
          <FunctionVersionRow
            key={version.version}
            env={envName}
            func={functionName}
            version={version}
            isActive={
              activeVersion && activeVersion.version === version.version
            }
          />
        ),
      })
    })
    return <Table columns={columns} rows={rows} />
  }
}

FunctionVersions.propTypes = {
  envName: PropTypes.string.isRequired,
  functionName: PropTypes.string.isRequired,
  func: PropTypes.object,
  activateFunctionTab: PropTypes.func,
}

export default connect(makeMapStateToProps, dispatchProps)(FunctionVersions)
