const CURRENT_RUNTIME_ENV = 'prod'
const PROD_API_HOST = 'https://api.example.com'

let localConfig = {}
try {
  localConfig = require('./config.local')
} catch (error) {
  localConfig = {}
}

function getLocalApiHost(envVersion) {
  if (localConfig && typeof localConfig.getApiHost === 'function') {
    return String(localConfig.getApiHost(envVersion) || '').trim()
  }
  if (localConfig && localConfig.apiHosts && typeof localConfig.apiHosts[envVersion] === 'string') {
    return localConfig.apiHosts[envVersion].trim()
  }
  if (envVersion === 'develop' && localConfig && typeof localConfig.apiHost === 'string') {
    return localConfig.apiHost.trim()
  }
  return ''
}

function getRuntimeEnvName() {
  return CURRENT_RUNTIME_ENV
}

function getApiHost() {
  const envVersion = getMiniProgramEnvVersion()
  return getLocalApiHost(envVersion) || PROD_API_HOST
}

function getUserApiBase() {
  return getApiHost() + '/api/v1/user'
}

function getMiniProgramEnvVersion() {
  try {
    const accountInfo = wx.getAccountInfoSync()
    return accountInfo && accountInfo.miniProgram && accountInfo.miniProgram.envVersion
      ? accountInfo.miniProgram.envVersion
      : 'release'
  } catch (error) {
    return 'release'
  }
}

function isDevelopVersion() {
  return getMiniProgramEnvVersion() === 'develop'
}

module.exports = {
  CURRENT_RUNTIME_ENV,
  getRuntimeEnvName,
  getApiHost,
  getUserApiBase,
  getMiniProgramEnvVersion,
  isDevelopVersion
}
