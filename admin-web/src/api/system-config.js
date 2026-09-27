import request from '../utils/request'

export function getSystemConfigs() {
  return request.get('/system-configs')
}

export function updateSystemConfig(data) {
  return request.put('/system-config', data)
}