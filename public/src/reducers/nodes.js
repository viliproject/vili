import { CHANGE_NODE } from '../constants'
import Node from '../models/Node'

import { getInitialState, changeObject } from './utils'

const initialState = getInitialState()

export default function (state = initialState, action) {
  switch (action.type) {
    case CHANGE_NODE:
      return changeObject(state, action, Node)
    default:
      return state
  }
}
