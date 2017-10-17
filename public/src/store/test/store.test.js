/* eslint-env mocha */
/* global expect */

import configureStore from '../index'

describe('Store', () => {
  it('should create redux store', () => {
    const store = configureStore()
    expect(store).to.exist
  })
})
