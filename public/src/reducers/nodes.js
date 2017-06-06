import { CHANGE_NODE } from '../constants'
import NodeModel from '../models/NodeModel'

import { getInitialState, changeObject } from './utils'

const initialState = getInitialState()

export default function (state = initialState, action) {
  switch (action.type) {
    case CHANGE_NODE:
      return changeObject(state, action, NodeModel)
    default:
      return state
  }
}
