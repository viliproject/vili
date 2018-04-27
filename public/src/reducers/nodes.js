import { CHANGE_NODE } from "../constants"
import Node from "../models/Node"

import { getInitialState, changeObject } from "./utils"

const initialState = getInitialState()

export default function(state = initialState, action) {
  switch (action.type) {
    case CHANGE_NODE:
      let { env, event: { type: eventType, object, list } } = action.payload
      if (eventType === "MODIFIED") {
        object = {
          ...object,
          metadata: {
            name: object.metadata.name,
          },
          status: {
            conditions: object.status.conditions,
          },
        }
        if (object.status.conditions) {
          object.status.conditions = object.status.conditions.map(c => {
            return {
              status: c.status,
              reason: c.reason,
              message: c.message,
            }
          })
        }
      }
      return changeObject(
        state,
        { payload: { env, event: { type: eventType, object, list } } },
        Node
      )
    default:
      return state
  }
}
