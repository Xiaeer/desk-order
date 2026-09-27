// Local environment overrides only.
// Copy this file as config.local.js and do not commit it.

module.exports = {
  environments: {
    local: { apiHost: 'http://127.0.0.1:8080' },
    prod: { apiHost: 'https://api.example.com' }
  }
}
