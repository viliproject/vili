import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import _ from "underscore"
import Immutable from "immutable"

import displayTime from "../../lib/displayTime"
import Table from "../../components/Table"
import Loading from "../../components/Loading"
import FunctionRow from "../../components/functions/FunctionRow"
import { activateFunctionTab } from "../../actions/app"
import { getFunctionRepository } from "../../actions/functions"

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
  getFunctionRepository,
}

export class Function extends React.Component {
  componentDidMount() {
    this.props.activateFunctionTab("home")
    this.fetchData()
  }

  componentDidUpdate(prevProps) {
    if (
      this.props.envName !== prevProps.envName ||
      this.props.functionName !== prevProps.functionName
    ) {
      this.fetchData()
    }
  }

  fetchData = () => {
    const { envName, functionName, getFunctionRepository } = this.props
    getFunctionRepository(envName, functionName)
  }

  render() {
    const { envName, functionName, func } = this.props
    if (!func || !func.get("repository")) {
      return <Loading />
    }

    const columns = [
      { title: "Tag", key: "tag", style: { width: "250px" } },
      { title: "Branch", key: "branch", style: { width: "200px" } },
      { title: "Revision", key: "revision", style: { width: "90px" } },
      { title: "Build Time", key: "buildTime", style: { width: "180px" } },
      { title: "Deployed", key: "deployedAt", style: { textAlign: "right" } },
      { title: "Actions", key: "actions", style: { textAlign: "right" } },
    ]

    let rows = []
    const funcVersions = func.get("object", {}).versions || Immutable.List()
    const activeVersion = func.get("object", {}).activeVersion
    func.get("repository").forEach(image => {
      const imageVersions = funcVersions.filter(
        v => v.get("tag") === image.get("tag")
      )
      const buildTime = new Date(image.get("lastModified"))
      rows.push({
        component: (
          <FunctionRow
            key={image.get("tag")}
            env={envName}
            functionName={functionName}
            tag={image.get("tag")}
            branch={image.get("branch")}
            revision={image.get("revision")}
            buildTime={displayTime(buildTime)}
            versions={imageVersions}
            activeVersion={activeVersion}
          />
        ),
        time: buildTime.getTime(),
      })
    })

    rows = _.sortBy(rows, row => {
      return -row.time
    })

    return <Table columns={columns} rows={rows} />
  }
}

Function.propTypes = {
  envName: PropTypes.string,
  functionName: PropTypes.string,
  activateFunctionTab: PropTypes.func,
  getFunctionRepository: PropTypes.func,
  func: PropTypes.object,
}

export default connect(makeMapStateToProps, dispatchProps)(Function)
