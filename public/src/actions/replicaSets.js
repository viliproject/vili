import { CHANGE_REPLICA_SET } from '../constants'

import { subObjects } from './utils'

export function subReplicaSets (env) {
  return subObjects(CHANGE_REPLICA_SET, 'replicaSets', env)
}
