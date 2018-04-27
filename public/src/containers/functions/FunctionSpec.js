import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"

import Loading from "../../components/Loading"
import { activateFunctionTab } from "../../actions/app"
import { getFunctionSpec } from "../../actions/functions"

function mapStateToProps(state, ownProps) {
  const { envName, functionName } = ownProps
  const func = state.functions.lookUpData(envName, functionName)
  return {
    func,
  }
}

const dispatchProps = {
  activateFunctionTab,
  getFunctionSpec,
}

export class FunctionSpec extends React.Component {
  componentDidMount() {
    this.props.activateFunctionTab("spec")
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
    const { envName, functionName, getFunctionSpec } = this.props
    getFunctionSpec(envName, functionName)
  }

  render() {
    const { func } = this.props
    if (!func || !func.get("spec")) {
      return <Loading />
    }
    return (
      <div className="col-md-8">
        <div id="source-yaml">
          <pre>
            <code>{func.get("spec")}</code>
          </pre>
        </div>
      </div>
    )
  }
}

FunctionSpec.propTypes = {
  envName: PropTypes.string,
  functionName: PropTypes.string,
  func: PropTypes.object,
  activateFunctionTab: PropTypes.func,
  getFunctionSpec: PropTypes.func,
}

export default connect(mapStateToProps, dispatchProps)(FunctionSpec)
