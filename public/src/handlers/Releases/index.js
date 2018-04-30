import PropTypes from "prop-types"
import React from "react"
import { Route, Switch } from "react-router"

import ReleasesList from "../../handlers/ReleasesList"
import ReleaseCreate from "../../handlers/ReleaseCreate"
import Release from "../../handlers/Release"
import ReleaseRollout from "../../handlers/ReleaseRollout"
import NotFoundPage from "../../components/NotFoundPage"

export class Releases extends React.Component {
  render() {
    const prefix = this.props.match.path
    return (
      <Switch>
        <Route exact path={`${prefix}`} component={ReleasesList} />
        <Route exact path={`${prefix}/create`} component={ReleaseCreate} />
        <Route exact path={`${prefix}/:release`} component={Release} />
        <Route
          path={`${prefix}/:release/rollouts/:rollout`}
          component={ReleaseRollout}
        />
        <Route component={NotFoundPage} />
      </Switch>
    )
  }
}

Releases.propTypes = {
  match: PropTypes.object.isRequired,
}

export default Releases
