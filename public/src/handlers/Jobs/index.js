import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import JobsList from "../../handlers/JobsList"
import JobContainer from "../../handlers/JobContainer"

export class Jobs extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={JobsList} />
        <Route path={`${prefix}/:job`} component={JobContainer} />
      </Switch>
    )
  }
}

Jobs.propTypes = {
  match: PropTypes.object.isRequired,
}

export default Jobs
