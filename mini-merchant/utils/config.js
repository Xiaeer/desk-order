const CURRENT_RUNTIME_ENV = 'prod' // local, prod

const ENVIRONMENTS = {
  local: {
    apiHost: 'http://127.0.0.1:8080'
  },
  prod: {
    apiHost: 'https://api.example.com'
  }
}

let localConfig = {}
try {
  localConfig = require('./config.local')
} catch (error) {
  localConfig = {}
}

function getRuntimeEnvName() {
  return ENVIRONMENTS[CURRENT_RUNTIME_ENV] ? CURRENT_RUNTIME_ENV : 'prod'
}

function getEnvConfig() {
  const envName = getRuntimeEnvName()
  const defaults = ENVIRONMENTS[envName] || ENVIRONMENTS.prod
  const overrides = localConfig && localConfig.environments && localConfig.environments[envName]
  return Object.assign({}, defaults, overrides || {})
}

function getApiHost() {
  return getEnvConfig().apiHost
}

function getMerchantApiBase() {
  return getApiHost() + '/api/v1/merchant'
}

function getUserApiBase() {
  return getApiHost() + '/api/v1/user'
}

module.exports = {
  CURRENT_RUNTIME_ENV,
  ENVIRONMENTS,
  getRuntimeEnvName,
  getApiHost,
  getMerchantApiBase,
  getUserApiBase
}
