// Local environment overrides only.
// 1) Copy this file as config.local.js
// 2) Set the API host for each WeChat environment
// 3) Do not commit config.local.js

module.exports = {
  apiHosts: {
    develop: 'http://127.0.0.1:8080',
    trial: 'https://api-trial.example.com',
    release: 'https://api.example.com'
  }
}
