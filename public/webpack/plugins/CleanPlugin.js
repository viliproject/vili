'use strict'
const rimraf = require('rimraf')

class CleanPlugin {
  constructor (path) {
    this.path = path
  }

  apply (compiler) {
    rimraf.sync(this.path)
  }
}

module.exports = CleanPlugin
