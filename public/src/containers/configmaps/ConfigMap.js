import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import { Table, Button, ButtonToolbar, Modal } from "react-bootstrap"
import Immutable from "immutable"

import Loading from "../../components/Loading"
import { activateNav } from "../../actions/app"
import {
  getConfigMapSpec,
  createConfigMap,
  deleteConfigMap,
} from "../../actions/configmaps"

import ConfigMapRow from "./ConfigMapRow"

function mapStateToProps(state, ownProps) {
  const { envName, configmapName } = ownProps
  const configmap = state.configmaps.lookUpData(envName, configmapName)
  return {
    configmap,
  }
}

const dispatchProps = {
  activateNav,
  getConfigMapSpec,
  createConfigMap,
  deleteConfigMap,
}

export class ConfigMap extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      showDeleteModal: false,
    }
  }

  componentDidMount() {
    this.props.activateNav("configmaps", this.props.configmapName)
    this.fetchData()
  }

  componentDidUpdate(prevProps) {
    if (
      this.props.envName !== prevProps.envName ||
      this.props.configmapName !== prevProps.configmapName
    ) {
      this.props.activateNav("configmaps", this.props.configmapName)
      this.fetchData()
    }
  }

  fetchData = () => {
    const { envName, configmapName, getConfigMapSpec } = this.props
    getConfigMapSpec(envName, configmapName)
  }

  clickCreate = () => {
    const { envName, configmapName, createConfigMap } = this.props
    createConfigMap(envName, configmapName)
  }

  clickDelete = () => {
    this.setState({
      showDeleteModal: true,
    })
  }

  get deleteModal() {
    const { configmapName } = this.props
    return (
      <Modal
        show={this.state.showDeleteModal}
        onHide={() => this.setState({ showDeleteModal: false })}
      >
        <Modal.Header>
          <Modal.Title>Delete {configmapName}?</Modal.Title>
        </Modal.Header>

        <Modal.Footer>
          <Button onClick={() => this.setState({ showDeleteModal: false })}>
            Close
          </Button>
          <Button onClick={this.deleteConfigMap} bsStyle="danger">
            Delete
          </Button>
        </Modal.Footer>
      </Modal>
    )
  }

  deleteConfigMap = () => {
    const { envName, configmapName, deleteConfigMap } = this.props
    deleteConfigMap(envName, configmapName)
    this.fetchData()
    this.setState({ showDeleteModal: false })
  }

  render() {
    const { envName, configmapName, configmap } = this.props
    const header = [
      <ol key="breadcrumb" className="breadcrumb">
        <li>
          <Link to={`/${envName}`}>{envName}</Link>
        </li>
        <li>
          <Link to={`/${envName}/configmaps`}>Config Maps</Link>
        </li>
        <li className="active">{configmapName}</li>
      </ol>,
    ]
    if (!configmap || !configmap.get("spec")) {
      return (
        <div>
          <div className="view-header">{header}</div>
          <Loading />
        </div>
      )
    }

    const buttons = []
    if (!configmap.get("object")) {
      buttons.push(
        <Button
          key="create"
          onClick={this.clickCreate}
          bsStyle="success"
          bsSize="small"
        >
          Create Config Map
        </Button>
      )
    } else {
      buttons.push(
        <Button
          key="delete"
          onClick={this.clickDelete}
          bsStyle="danger"
          bsSize="small"
        >
          Delete Config Map
        </Button>
      )
    }
    header.push(
      <ButtonToolbar key="toolbar" bsClass="pull-right">
        {buttons}
      </ButtonToolbar>
    )

    const actualMap = configmap.getIn(["object", "data"], Immutable.Map())
    const specMap = configmap.getIn(["spec", "data"], Immutable.Map())

    const keys = Immutable.Set()
      .union(actualMap.keys(), specMap.keys())
      .sort()

    const rows = []
    keys.forEach(key => {
      rows.push(
        <ConfigMapRow
          key={key}
          keyName={key}
          envName={envName}
          configmapName={configmapName}
          val={actualMap.get(key)}
          specVal={specMap.get(key)}
          hasConfigMap={!!configmap.get("object")}
        />
      )
    })

    return (
      <div>
        <div className="view-header">{header}</div>
        {this.deleteModal}
        <div>
          <div className="text-danger">− Spec Value</div>
          <div className="text-success">+ Actual Value</div>
        </div>
        <Table>
          <thead>
            <tr>
              <th>Key</th>
              <th>Value</th>
              <th style={{ width: "200px" }}>Actions</th>
            </tr>
          </thead>
          <tbody>{rows}</tbody>
        </Table>
      </div>
    )
  }
}

ConfigMap.propTypes = {
  envName: PropTypes.string.isRequired,
  configmapName: PropTypes.string.isRequired,
  configmap: PropTypes.object,
  activateNav: PropTypes.func.isRequired,
  getConfigMapSpec: PropTypes.func.isRequired,
  createConfigMap: PropTypes.func.isRequired,
  deleteConfigMap: PropTypes.func.isRequired,
}

export default connect(mapStateToProps, dispatchProps)(ConfigMap)
