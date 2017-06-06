/* eslint-env mocha */
/* global expect */

import * as Actions from '../index'
import * as Constants from '../../constants'

describe('Global Actions', () => {
  it('should create valid appResize action', () => {
    const action = Actions.appResize(100, 100)
    expect(action.type).to.equal(Constants.APP_RESIZE)
    expect(action.payload.width).to.equal(100)
    expect(action.payload.height).to.equal(100)
  })
})
