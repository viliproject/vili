import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"

import Loading from "../../components/Loading"
import PodLog from "../../components/PodLog"
import { activateNav } from "../../actions/app"
import { subPodLog, unsubPodLog } from "../../actions/pods"

function mapStateToProps(state, ownProps) {
  const { envName, podName } = ownProps
  const pod = state.pods.lookUpData(envName, podName)
  return {
    pod,
  }
}

const dispatchProps = {
  activateNav,
  subPodLog,
  unsubPodLog,
}

export class Pod extends React.Component {
  componentDidMount() {
    this.props.activateNav("pods")
    this.subData()
  }

  componentDidUpdate(prevProps) {
    const { envName, podName } = this.props
    const { envName: prevEnvName, podName: prevPodName } = prevProps
    if (envName !== prevEnvName || podName !== prevPodName) {
      this.unsubData()
      this.subData()
    }
  }

  subData = () => {
    const { envName, podName, subPodLog } = this.props
    subPodLog(envName, podName)
  }

  unsubData = () => {
    const { envName, podName, unsubPodLog } = this.props
    unsubPodLog(envName, podName)
  }

  render() {
    const { envName, podName, pod: podData } = this.props
    const pod = podData && podData.get("object")
    const header = (
      <div className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li>
            <Link to={`/${envName}/pods`}>Pods</Link>
          </li>
          <li className="active">{podName}</li>
        </ol>
      </div>
    )
    if (!pod) {
      return (
        <div>
          {header}
          <Loading />
        </div>
      )
    }

    const metadata = [
      <dt key="title-ip">IP</dt>,
      <dd key="data-ip">{pod.getIn(["status", "podIP"])}</dd>,
      <dt key="title-phase">Phase</dt>,
      <dd key="data-phase">{pod.getIn(["status", "phase"])}</dd>,
      <dt key="title-node">Node</dt>,
      <dd key="data-node">
        <Link to={`/${envName}/nodes/${pod.getIn(["spec", "nodeName"])}`}>
          {pod.getIn(["spec", "nodeName"])}
        </Link>
      </dd>,
    ]
    if (pod.getLabel("app")) {
      metadata.push(<dt key="title-deployment">Deployment</dt>)
      metadata.push(
        <dd key="data-deployment">
          <Link to={`/${envName}/deployments/${pod.getLabel("app")}`}>
            {pod.getLabel("app")}
          </Link>
        </dd>
      )
    }
    if (pod.getLabel("job")) {
      metadata.push(<dt key="title-job">Job</dt>)
      metadata.push(
        <dd key="data-job">
          <Link to={`/${envName}/jobs/${pod.getLabel("job")}`}>
            {pod.getLabel("job")}
          </Link>
        </dd>
      )
    }
    if (pod.deployedBy) {
      metadata.push(<dt key="title-deployedBy">Deployed By</dt>)
      metadata.push(<dd key="data-deployedBy">{pod.deployedBy}</dd>)
    }
    return (
      <div>
        {header}
        <div>
          <h4>Metadata</h4>
          <dl className="dl-horizontal">{metadata}</dl>
        </div>
        <PodLog log={podData.get("log")} />
      </div>
    )
  }
}

Pod.propTypes = {
  envName: PropTypes.string,
  podName: PropTypes.string,
  pod: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  subPodLog: PropTypes.func.isRequired,
  unsubPodLog: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(Pod)
