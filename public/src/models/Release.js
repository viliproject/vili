import Immutable from 'immutable'

import displayTime from '../lib/displayTime'
import ReleaseRollout from './ReleaseRollout'

export default class Release extends Immutable.Record({
  targetEnv: undefined,
  name: undefined,
  link: undefined,
  waves: Immutable.List(),
  createdAt: undefined,
  createdBy: undefined,
  rollouts: Immutable.List()
}) {
  get createdAtHumanize () {
    return displayTime(new Date(this.get('createdAt')))
  }

  envRollouts (env) {
    return this
      .get('rollouts', Immutable.List())
      .filter((r) => r.get('env') === env)
      .map((r) => new ReleaseRollout(r))
      .sortBy((r) => -r.get('rolloutAtDate'))
  }
}
