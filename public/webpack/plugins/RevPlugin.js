'use strict'
const fs = require('fs')
const notifier = require('node-notifier')

class RevPlugin {
  constructor (options) {
    if (typeof options === 'string') {
      options = {
        path: options
      }
    }
    this.options = options || {}
  }

  apply (compiler) {
    compiler.plugin('done', (stats) => {
      const data = stats.toJson({ hash: true })
      fs.writeFileSync(
          this.options.path,
          JSON.stringify({
            hash: data.hash,
            assets: data.assets
          })
      )

      let message = 'BUILT!!!'
      if (data.errors && data.errors.length > 0) {
        message = 'FAILED!!!'
      }

      notifier.notify({
        'title': this.options.notifyTitle || 'APP',
        'message': message
      })
    })
  }
}

module.exports = RevPlugin
