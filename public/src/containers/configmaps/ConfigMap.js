import PropTypes from 'prop-types'
import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { Table, Button, ButtonToolbar, Modal, FormGroup, FormControl, ControlLabel } from 'react-bootstrap'
import _ from 'underscore'
import Immutable from 'immutable'

import Loading from '../../components/Loading'
import { activateNav } from '../../actions/app'
import { getConfigMapSpec, createConfigMap, deleteConfigMap, setConfigMapKeys, deleteConfigMapKey } from '../../actions/configmaps'

function mapStateToProps (state, ownProps) {
  const { env, configmap: configmapName } = ownProps.params
  const configmap = state.configmaps.lookUpData(env, configmapName)
  return {
    configmap
  }
}

const dispatchProps = {
  activateNav,
  getConfigMapSpec,
  createConfigMap,
  deleteConfigMap
}

export class ConfigMap extends React.Component {
  constructor (props) {
    super(props)
    this.state = {
      showDeleteModal: false
    }
  }

  componentDidMount () {
    this.props.activateNav('configmaps', this.props.params.configmap)
    this.fetchData()
  }

  componentDidUpdate (prevProps) {
    if (this.props.params !== prevProps.params) {
      this.props.activateNav('configmaps', this.props.params.configmap)
      this.fetchData()
    }
  }

  fetchData = () => {
    const { params, getConfigMapSpec } = this.props
    getConfigMapSpec(params.env, params.configmap)
  }

  clickCreate = () => {
    const { params, createConfigMap } = this.props
    createConfigMap(params.env, params.configmap)
  }

  clickDelete = () => {
    this.setState({
      showDeleteModal: true
    })
  }

  get deleteModal () {
    const { params } = this.props
    return (
      <Modal show={this.state.showDeleteModal} onHide={() => this.setState({ showDeleteModal: false })}>
        <Modal.Header>
          <Modal.Title>Delete {params.configmap}?</Modal.Title>
        </Modal.Header>

        <Modal.Footer>
          <Button onClick={() => this.setState({ showDeleteModal: false })}>Close</Button>
          <Button onClick={this.deleteConfigMap} bsStyle='danger'>Delete</Button>
        </Modal.Footer>
      </Modal>
    )
  }

  deleteConfigMap = () => {
    const { params } = this.props
    this.props.deleteConfigMap(params.env, params.configmap)
    this.fetchData()
    this.setState({ showDeleteModal: false })
  }

  render () {
    const { params, configmap } = this.props
    const header = [
      <ol key='breadcrumb' className='breadcrumb'>
        <li><Link to={`/${params.env}`}>{params.env}</Link></li>
        <li><Link to={`/${params.env}/configmaps`}>Config Maps</Link></li>
        <li className='active'>{params.configmap}</li>
      </ol>
    ]
    if (!configmap || !configmap.get('spec')) {
      return (
        <div>
          <div className='view-header'>{header}</div>
          <Loading />
        </div>
      )
    }

    const buttons = []
    if (!configmap.get('object')) {
      buttons.push(<Button key='create' onClick={this.clickCreate} bsStyle='success' bsSize='small'>Create Config Map</Button>)
    } else {
      buttons.push(<Button key='delete' onClick={this.clickDelete} bsStyle='danger' bsSize='small'>Delete Config Map</Button>)
    }
    header.push(<ButtonToolbar key='toolbar' bsClass='pull-right'>{buttons}</ButtonToolbar>)

    const actualMap = configmap.getIn(['object', 'data'], Immutable.Map())
    const specMap = configmap.getIn(['spec', 'data'], Immutable.Map())

    const keys = Immutable.Set().union(actualMap.keys(), specMap.keys()).sort()

    const rows = []
    keys.forEach((key) => {
      rows.push(
        <Row
          key={key}
          keyName={key}
          env={params.env}
          configmapName={params.configmap}
          val={actualMap.get(key)}
          specVal={specMap.get(key)}
          hasConfigMap={!!configmap.get('object')}
        />
      )
    })

    return (
      <div>
        <div className='view-header'>{header}</div>
        {this.deleteModal}
        <div>
          <div className='text-danger'>− Spec Value</div>
          <div className='text-success'>+ Actual Value</div>
        </div>
        <Table>
          <thead>
            <tr>
              <th>Key</th>
              <th>Value</th>
              <th style={{width: '200px'}}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </Table>
      </div>
    )
  }
}

ConfigMap.propTypes = {
  activateNav: PropTypes.func.isRequired,
  getConfigMapSpec: PropTypes.func.isRequired,
  createConfigMap: PropTypes.func.isRequired,
  deleteConfigMap: PropTypes.func.isRequired,
  params: PropTypes.object,
  location: PropTypes.object,
  configmap: PropTypes.object
}

export default connect(mapStateToProps, dispatchProps)(ConfigMap)

const rowDispatchProps = {
  setConfigMapKeys,
  deleteConfigMapKey
}

@connect(null, rowDispatchProps)
class Row extends React.Component {
  static propTypes = {
    setConfigMapKeys: PropTypes.func.isRequired,
    deleteConfigMapKey: PropTypes.func.isRequired,
    env: PropTypes.string,
    configmapName: PropTypes.string,
    keyName: PropTypes.string,
    val: PropTypes.string,
    specVal: PropTypes.string,
    hasConfigMap: PropTypes.bool
  }

  constructor (props) {
    super(props)
    this.state = {
      showEditModal: false,
      showDeleteModal: false
    }
  }

  clickReset = () => {
    const { specVal } = this.props
    this.setState({
      showEditModal: true,
      editValue: specVal
    })
  }

  clickEdit = () => {
    const { val } = this.props
    this.setState({
      showEditModal: true,
      editValue: val
    })
  }

  get editModal () {
    const { keyName, val } = this.props
    const edited = val !== this.state.editValue
    return (
      <Modal show={this.state.showEditModal} onHide={() => this.setState({ showEditModal: false })}>
        <Modal.Body>
          <form>
            <FormGroup validationState={edited ? 'success' : null}>
              <ControlLabel>{keyName}</ControlLabel>
              <FormControl
                type='text'
                placeholder='Value'
                value={this.state.editValue}
                onChange={this.handleEditModalChange}
              />
              <FormControl.Feedback />
            </FormGroup>
          </form>

        </Modal.Body>

        <Modal.Footer>
          <Button onClick={() => this.setState({ showEditModal: false })}>Close</Button>
          <Button onClick={this.saveEditModal} bsStyle='primary' disabled={!edited}>Save</Button>
        </Modal.Footer>
      </Modal>
    )
  }

  handleEditModalChange = (e) => {
    this.setState({ editValue: e.target.value })
  }

  saveEditModal = () => {
    const { env, configmapName, keyName } = this.props
    this.props.setConfigMapKeys(env, configmapName, {[keyName]: this.state.editValue})
    this.setState({ showEditModal: false })
  }

  clickDelete = () => {
    this.setState({
      showDeleteModal: true
    })
  }

  get deleteModal () {
    const { keyName } = this.props
    return (
      <Modal show={this.state.showDeleteModal} onHide={() => this.setState({ showDeleteModal: false })}>
        <Modal.Header>
          <Modal.Title>Delete {keyName}?</Modal.Title>
        </Modal.Header>

        <Modal.Footer>
          <Button onClick={() => this.setState({ showDeleteModal: false })}>Close</Button>
          <Button onClick={this.saveDeleteModal} bsStyle='danger'>Delete</Button>
        </Modal.Footer>
      </Modal>
    )
  }

  saveDeleteModal = () => {
    const { env, configmapName, keyName } = this.props
    this.props.deleteConfigMapKey(env, configmapName, keyName)
    this.setState({ showDeleteModal: false })
  }

  render () {
    const { keyName, val, specVal, hasConfigMap } = this.props

    const value = []
    const actions = []
    const style = {marginLeft: '10px'}
    if (!_.isString(val)) {
      value.push(<div key='specVal' className='text-danger'>− {specVal}</div>)
      actions.push(<Button key='reset' onClick={this.clickReset} style={style} bsStyle='success' bsSize='xs'>Add</Button>)
    } else if (!_.isString(specVal)) {
      value.push(<div key='val' className='text-success'>+ {val}</div>)
      actions.push(<Button key='delete' onClick={this.clickDelete} style={style} bsStyle='warning' bsSize='xs'>Delete</Button>)
      actions.push(<Button key='edit' onClick={this.clickEdit} style={style} bsStyle='primary' bsSize='xs'>Edit</Button>)
    } else if (val !== specVal) {
      value.push(<div key='specVal' className='text-danger'>− {specVal}</div>)
      value.push(<div key='val' className='text-success'>+ {val}</div>)
      actions.push(<Button key='reset' onClick={this.clickReset} style={style} bsStyle='warning' bsSize='xs'>Reset</Button>)
      actions.push(<Button key='delete' onClick={this.clickDelete} style={style} bsStyle='danger' bsSize='xs'>Delete</Button>)
      actions.push(<Button key='edit' onClick={this.clickEdit} style={style} bsStyle='primary' bsSize='xs'>Edit</Button>)
    } else {
      value.push(<div key='val'>{val}</div>)
      actions.push(<Button key='delete' onClick={this.clickDelete} style={style} bsStyle='danger' bsSize='xs'>Delete</Button>)
      actions.push(<Button key='edit' onClick={this.clickEdit} style={style} bsStyle='primary' bsSize='xs'>Edit</Button>)
    }
    return (
      <tr key={keyName}>
        <th>{keyName}</th>
        <td>{value}</td>
        <td><div style={{textAlign: 'right'}}>{hasConfigMap ? actions : ''}</div></td>
        {this.editModal}
        {this.deleteModal}
      </tr>
    )
  }
}
