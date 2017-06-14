import PropTypes from 'prop-types'
import React from 'react'
import { Nav } from 'react-bootstrap'
import _ from 'underscore'

import LinkMenuItem from '../LinkMenuItem'

export default class SideNav extends React.Component {
  static propTypes = {
    env: PropTypes.object,
    nav: PropTypes.object
  }

  get navItems () {
    const { env, nav } = this.props
    if (!env || !nav) {
      return
    }
    var items = []

    items.push(
      <LinkMenuItem key='releases' to={`/${env.name}/releases`}
        active={nav.item === 'releases'}
      >
        Releases
      </LinkMenuItem>
    )

    if (!_.isEmpty(env.deployments)) {
      items.push(
        <LinkMenuItem key='deployments' to={`/${env.name}/deployments`}
          active={nav.item === 'deployments' && !nav.subItem}
        >
          Deployments
        </LinkMenuItem>)
      if (nav.item === 'deployments') {
        _.map(env.deployments, function (deployment) {
          items.push(
            <LinkMenuItem key={`deployments-${deployment}`} to={`/${env.name}/deployments/${deployment}`} subitem
              active={nav.item === 'deployments' && nav.subItem === deployment}
            >
              {deployment}
            </LinkMenuItem>)
        })
      }
    }
    if (!_.isEmpty(env.jobs)) {
      items.push(
        <LinkMenuItem key='jobs' to={`/${env.name}/jobs`}
          active={nav.item === 'jobs' && !nav.subItem}
        >
          Jobs
        </LinkMenuItem>)
      if (nav.item === 'jobs') {
        _.map(env.jobs, function (job) {
          items.push(
            <LinkMenuItem key={`jobs-${job}`} to={`/${env.name}/jobs/${job}`} subitem
              active={nav.item === 'jobs' && nav.subItem === job}
            >
              {job}
            </LinkMenuItem>)
        })
      }
    }
    if (!_.isEmpty(env.configmaps)) {
      items.push(
        <LinkMenuItem key='configmaps' to={`/${env.name}/configmaps`}
          active={nav.item === 'configmaps' && !nav.subItem}
        >Config Maps</LinkMenuItem>
      )
      if (nav.item === 'configmaps') {
        _.map(env.configmaps, function (configmap) {
          items.push(
            <LinkMenuItem key={`configmaps-${configmap}`} to={`/${env.name}/configmaps/${configmap}`} subitem
              active={nav.item === 'configmaps' && nav.subItem === configmap}
            >
              {configmap}
            </LinkMenuItem>)
        })
      }
    }
    items.push(
      <LinkMenuItem key='nodes' to={`/${env.name}/nodes`}
        active={nav.item === 'nodes'}
      >Nodes</LinkMenuItem>
    )
    items.push(
      <LinkMenuItem key='pods' to={`/${env.name}/pods`}
        active={nav.item === 'pods'}
      >Pods</LinkMenuItem>
    )
    return items
  }

  render () {
    return (
      <Nav className='side-nav' stacked>
        {this.navItems}
      </Nav>
    )
  }
}
