import React from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Link } from "react-router-dom"

import PodLog from "../../../components/PodLog"
import { subPodLog, unsubPodLog } from "../../../actions/pods"

function mapStateToProps(state, ownProps) {
  const { envName, podName } = ownProps
  const pod = state.pods.lookUpData(envName, podName)
  return {
    log: pod.get("log"),
  }
}

const dispatchProps = {
  subPodLog,
  unsubPodLog,
}

export class JobRunPod extends React.Component {
  componentDidMount() {
    this.subData()
  }

  componentWillUnmount() {
    this.unsubData()
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
    const { envName, podName, log } = this.props
    return (
      <div key={podName}>
        <h4>
          <Link to={`/${envName}/pods/${podName}`}>{podName}</Link>
        </h4>
        <PodLog log={log} />
      </div>
    )
  }
}

JobRunPod.propTypes = {
  envName: PropTypes.string,
  podName: PropTypes.string,
  log: PropTypes.string,
  subPodLog: PropTypes.func.isRequired,
  unsubPodLog: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(JobRunPod)
