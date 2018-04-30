import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import { Button } from "react-bootstrap"
import Immutable from "immutable"

import Table from "../../components/Table"
import { activateNav } from "../../actions/app"
import { setNodeSchedulable } from "../../actions/nodes"
import { makeLookUpObjects } from "../../selectors"

function makeMapStateToProps() {
  const lookUpObjects = makeLookUpObjects()
  return (state, ownProps) => {
    const { envName } = ownProps
    const nodes = lookUpObjects(state.nodes, envName)
    return {
      nodes,
    }
  }
}

const dispatchProps = {
  activateNav,
  setNodeSchedulable,
}

export class NodesList extends React.Component {
  componentDidMount() {
    this.props.activateNav("nodes")
  }

  setNodeSchedulable = (nodeName, action) => {
    const { envName, setNodeSchedulable } = this.props
    setNodeSchedulable(envName, nodeName, action)
  }

  render() {
    const { envName, nodes } = this.props
    const header = (
      <div className="view-header">
        <ol className="breadcrumb">
          <li>
            <Link to={`/${envName}`}>{envName}</Link>
          </li>
          <li className="active">Nodes</li>
        </ol>
      </div>
    )
    const columns = [
      { title: "Host", key: "host" },
      { title: "Instance Type", key: "instance_type" },
      { title: "Role", key: "role" },
      {
        title: "Capacity",
        key: "capacity",
        subcolumns: [
          { title: "CPU", key: "cpu_capacity" },
          { title: "Memory", key: "memory_capacity" },
          { title: "Pods", key: "pods_capacity" },
        ],
      },
      {
        title: "Versions",
        key: "versions",
        subcolumns: [
          { title: "CoreOS", key: "os_version" },
          { title: "Kubelet", key: "kubelet_version" },
          { title: "Proxy", key: "proxy_version" },
        ],
      },
      { title: "Created", key: "created" },
      { title: "Status", key: "status" },
      { title: "Actions", key: "actions" },
    ]

    const rows = []
    nodes.forEach(node => {
      var name = node.getIn(["metadata", "name"])
      var nodeStatuses = []
      node
        .getIn(["status", "conditions"], Immutable.List())
        .forEach(condition => {
          if (condition.get("status") === "True") {
            nodeStatuses.push(condition.get("type"))
          }
        })
      var actions
      if (node.getIn(["spec", "unschedulable"]) === true) {
        actions = (
          <Button
            bsStyle="success"
            bsSize="xs"
            onClick={() => this.setNodeSchedulable(name, "enable")}
          >
            Enable
          </Button>
        )
        nodeStatuses.push("Disabled")
      } else {
        actions = (
          <Button
            bsStyle="danger"
            bsSize="xs"
            onClick={() => this.setNodeSchedulable(name, "disable")}
          >
            Disable
          </Button>
        )
      }

      rows.push({
        host: <Link to={`/${envName}/nodes/${name}`}>{name}</Link>,
        instance_type: node.getLabel("beta.kubernetes.io/instance-type"),
        role: node.getLabel("airware.io/role"),
        cpu_capacity: node.getIn(["status", "capacity", "cpu"]),
        memory_capacity: node.memory,
        pods_capacity: node.getIn(["status", "capacity", "pods"]),
        os_version: node.getIn(["status", "nodeInfo", "osImage"]),
        kubelet_version: node.getIn(["status", "nodeInfo", "kubeletVersion"]),
        proxy_version: node.getIn(["status", "nodeInfo", "kubeProxyVersion"]),
        created: node.createdAt,
        status: nodeStatuses.join(","),
        actions: actions,
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

NodesList.propTypes = {
  envName: PropTypes.string.isRequired,
  nodes: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  setNodeSchedulable: PropTypes.func.isRequired,
}

export default connect(makeMapStateToProps, dispatchProps)(NodesList)
