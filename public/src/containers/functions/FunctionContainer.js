import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import _ from "underscore"

import { activateNav } from "../../actions/app"

const tabs = {
  home: "Home",
  versions: "Versions",
  spec: "Spec",
}

function mapStateToProps(state) {
  return {
    app: state.app,
  }
}

const dispatchProps = {
  activateNav,
}

export class FunctionContainer extends React.Component {
  componentDidMount() {
    const { functionName, activateNav } = this.props
    activateNav("functions", functionName)
  }

  componentDidUpdate(prevProps) {
    const { functionName, activateNav } = this.props
    if (functionName !== prevProps.functionName) {
      activateNav("functions", functionName)
    }
  }

  render() {
    const { envName, functionName, app, children } = this.props
    const tabElements = _.map(tabs, (name, key) => {
      let className = ""
      if (app.get("functionTab") === key) {
        className = "active"
      }
      let link = `/${envName}/functions/${functionName}`
      if (key !== "home") {
        link += `/${key}`
      }
      return (
        <li key={key} role="presentation" className={className}>
          <Link to={link}>{name}</Link>
        </li>
      )
    })
    return (
      <div>
        <div key="view-header" className="view-header">
          <ol className="breadcrumb">
            <li>
              <Link to={`/${envName}`}>{envName}</Link>
            </li>
            <li>
              <Link to={`/${envName}/functions`}>Functions</Link>
            </li>
            <li className="active">{functionName}</li>
          </ol>
          <ul className="nav nav-pills pull-right">{tabElements}</ul>
        </div>
        {children}
      </div>
    )
  }
}

FunctionContainer.propTypes = {
  envName: PropTypes.string,
  functionName: PropTypes.string,
  app: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  children: PropTypes.node,
}

export default connect(mapStateToProps, dispatchProps)(FunctionContainer)
