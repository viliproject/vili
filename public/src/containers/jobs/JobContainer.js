import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import _ from "underscore"

import { activateNav } from "../../actions/app"

const tabs = {
  home: "Home",
  runs: "Runs",
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

export class JobContainer extends React.Component {
  componentDidMount() {
    const { jobName, activateNav } = this.props
    activateNav("jobs", jobName)
  }

  componentDidUpdate(prevProps) {
    const { jobName, activateNav } = this.props
    if (jobName !== prevProps.jobName) {
      activateNav("jobs", jobName)
    }
  }

  render() {
    const { envName, jobName, app, children } = this.props
    const tabElements = _.map(tabs, (name, key) => {
      let className = ""
      if (app.get("jobTab") === key) {
        className = "active"
      }
      let link = `/${envName}/jobs/${jobName}`
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
              <Link to={`/${envName}/jobs`}>Jobs</Link>
            </li>
            <li className="active">{jobName}</li>
          </ol>
          <ul className="nav nav-pills pull-right">{tabElements}</ul>
        </div>
        {children}
      </div>
    )
  }
}

JobContainer.propTypes = {
  envName: PropTypes.string,
  jobName: PropTypes.string,
  app: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  children: PropTypes.node,
}

export default connect(mapStateToProps, dispatchProps)(JobContainer)
