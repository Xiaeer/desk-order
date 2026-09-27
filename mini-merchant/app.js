const { getRuntimeEnvName, getApiHost } = require('./utils/config')

App({
	onLaunch() {
		console.log('[DeskOrder][mini-merchant] runtime_env=', getRuntimeEnvName(), 'api_host=', getApiHost())
	}
})
