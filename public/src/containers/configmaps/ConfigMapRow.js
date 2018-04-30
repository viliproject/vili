import PropTypes from "prop-types"
import React from "react"
import { connect } from "react-redux"
import {
  Button,
  Modal,
  FormGroup,
  FormControl,
  ControlLabel,
} from "react-bootstrap"
import _ from "underscore"

import { setConfigMapKeys, deleteConfigMapKey } from "../../actions/configmaps"

const dispatchProps = {
  setConfigMapKeys,
  deleteConfigMapKey,
}

export class ConfigMapRow extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      showEditModal: false,
      showDeleteModal: false,
    }
  }

  clickReset = () => {
    const { specVal } = this.props
    this.setState({
      showEditModal: true,
      editValue: specVal,
    })
  }

  clickEdit = () => {
    const { val } = this.props
    this.setState({
      showEditModal: true,
      editValue: val,
    })
  }

  get editModal() {
    const { keyName, val } = this.props
    const edited = val !== this.state.editValue
    return (
      <Modal
        show={this.state.showEditModal}
        onHide={() => this.setState({ showEditModal: false })}
      >
        <Modal.Body>
          <form>
            <FormGroup validationState={edited ? "success" : null}>
              <ControlLabel>{keyName}</ControlLabel>
              <FormControl
                type="text"
                placeholder="Value"
                value={this.state.editValue}
                onChange={this.handleEditModalChange}
              />
              <FormControl.Feedback />
            </FormGroup>
          </form>
        </Modal.Body>

        <Modal.Footer>
          <Button onClick={() => this.setState({ showEditModal: false })}>
            Close
          </Button>
          <Button
            onClick={this.saveEditModal}
            bsStyle="primary"
            disabled={!edited}
          >
            Save
          </Button>
        </Modal.Footer>
      </Modal>
    )
  }

  handleEditModalChange = e => {
    this.setState({ editValue: e.target.value })
  }

  saveEditModal = () => {
    const { envName, configmapName, keyName, setConfigMapKeys } = this.props
    setConfigMapKeys(envName, configmapName, {
      [keyName]: this.state.editValue,
    })
    this.setState({ showEditModal: false })
  }

  clickDelete = () => {
    this.setState({
      showDeleteModal: true,
    })
  }

  get deleteModal() {
    const { keyName } = this.props
    return (
      <Modal
        show={this.state.showDeleteModal}
        onHide={() => this.setState({ showDeleteModal: false })}
      >
        <Modal.Header>
          <Modal.Title>Delete {keyName}?</Modal.Title>
        </Modal.Header>

        <Modal.Footer>
          <Button onClick={() => this.setState({ showDeleteModal: false })}>
            Close
          </Button>
          <Button onClick={this.saveDeleteModal} bsStyle="danger">
            Delete
          </Button>
        </Modal.Footer>
      </Modal>
    )
  }

  saveDeleteModal = () => {
    const { envName, configmapName, keyName, deleteConfigMapKey } = this.props
    deleteConfigMapKey(envName, configmapName, keyName)
    this.setState({ showDeleteModal: false })
  }

  render() {
    const { keyName, val, specVal, hasConfigMap } = this.props

    const value = []
    const actions = []
    const style = { marginLeft: "10px" }
    if (!_.isString(val)) {
      value.push(
        <div key="specVal" className="text-danger">
          − {specVal}
        </div>
      )
      actions.push(
        <Button
          key="reset"
          onClick={this.clickReset}
          style={style}
          bsStyle="success"
          bsSize="xs"
        >
          Add
        </Button>
      )
    } else if (!_.isString(specVal)) {
      value.push(
        <div key="val" className="text-success">
          + {val}
        </div>
      )
      actions.push(
        <Button
          key="delete"
          onClick={this.clickDelete}
          style={style}
          bsStyle="warning"
          bsSize="xs"
        >
          Delete
        </Button>
      )
      actions.push(
        <Button
          key="edit"
          onClick={this.clickEdit}
          style={style}
          bsStyle="primary"
          bsSize="xs"
        >
          Edit
        </Button>
      )
    } else if (val !== specVal) {
      value.push(
        <div key="specVal" className="text-danger">
          − {specVal}
        </div>
      )
      value.push(
        <div key="val" className="text-success">
          + {val}
        </div>
      )
      actions.push(
        <Button
          key="reset"
          onClick={this.clickReset}
          style={style}
          bsStyle="warning"
          bsSize="xs"
        >
          Reset
        </Button>
      )
      actions.push(
        <Button
          key="delete"
          onClick={this.clickDelete}
          style={style}
          bsStyle="danger"
          bsSize="xs"
        >
          Delete
        </Button>
      )
      actions.push(
        <Button
          key="edit"
          onClick={this.clickEdit}
          style={style}
          bsStyle="primary"
          bsSize="xs"
        >
          Edit
        </Button>
      )
    } else {
      value.push(<div key="val">{val}</div>)
      actions.push(
        <Button
          key="delete"
          onClick={this.clickDelete}
          style={style}
          bsStyle="danger"
          bsSize="xs"
        >
          Delete
        </Button>
      )
      actions.push(
        <Button
          key="edit"
          onClick={this.clickEdit}
          style={style}
          bsStyle="primary"
          bsSize="xs"
        >
          Edit
        </Button>
      )
    }
    return (
      <tr key={keyName}>
        <th>{keyName}</th>
        <td>{value}</td>
        <td>
          <div style={{ textAlign: "right" }}>
            {hasConfigMap ? actions : ""}
          </div>
        </td>
        {this.editModal}
        {this.deleteModal}
      </tr>
    )
  }
}

ConfigMapRow.propTypes = {
  setConfigMapKeys: PropTypes.func.isRequired,
  deleteConfigMapKey: PropTypes.func.isRequired,
  envName: PropTypes.string,
  configmapName: PropTypes.string,
  keyName: PropTypes.string,
  val: PropTypes.string,
  specVal: PropTypes.string,
  hasConfigMap: PropTypes.bool,
}

export default connect(null, dispatchProps)(ConfigMapRow)
