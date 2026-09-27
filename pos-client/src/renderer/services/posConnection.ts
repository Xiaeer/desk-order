export function createPOSConnectionService({
	state,
	render,
	hasConfig,
	buildWsUrl,
	createPOSSocket,
	pingPOSLogin,
	formatTime,
	onMessage,
}) {
	let reconnectAttempts = 0;

	function nextReconnectDelay(): number {
		const cappedAttempt = Math.min(reconnectAttempts, 6);
		const baseDelay = 2000;
		const maxDelay = 30000;
		const exponentDelay = baseDelay * (2 ** cappedAttempt);
		const jitter = Math.floor(Math.random() * 500);
		return Math.min(maxDelay, exponentDelay) + jitter;
	}

	function destroySocket() {
		if (state.socketClient) {
			state.socketClient.close();
			state.socketClient = null;
		}
	}

	function clearReconnectTimer() {
		if (state.reconnectTimer) {
			window.clearTimeout(state.reconnectTimer);
			state.reconnectTimer = null;
		}
	}

	function scheduleReconnect() {
		if (state.reconnectTimer || !hasConfig()) {
			return;
		}
		const delay = nextReconnectDelay();
		reconnectAttempts += 1;
		state.connectionMessage = `连接异常，${Math.ceil(delay / 1000)} 秒后重试`;
		render();
		state.reconnectTimer = window.setTimeout(() => {
			state.reconnectTimer = null;
			connectPOS(true);
		}, delay);
	}

	async function connectPOS(forceReconnect = false) {
		if (!hasConfig()) {
			state.connectionStatus = 'idle';
			state.connectionMessage = '请先绑定店铺';
			render();
			return;
		}
		if (forceReconnect) {
			clearReconnectTimer();
			destroySocket();
			reconnectAttempts = 0;
		}
		if (state.socketClient) {
			return;
		}
		state.connectionStatus = 'connecting';
		state.connectionMessage = '正在连接 POS 推单通道';
		render();

		try {
			await pingPOSLogin(state.config);
		} catch (error) {
			state.connectionStatus = 'error';
			state.connectionMessage = error.message;
			render();
			scheduleReconnect();
			return;
		}

		try {
			state.socketClient = createPOSSocket({
				url: buildWsUrl(state.config.apiBase, state.config.shopId, state.config.posToken),
				onOpen: () => {
					reconnectAttempts = 0;
					state.connectionStatus = 'connected';
					state.connectionMessage = '实时推单已连接';
					state.lastEventAt = formatTime(new Date().toISOString());
					render();
				},
				onClose: (details: { opened?: boolean } = {}) => {
					state.socketClient = null;
					if (state.connectionStatus !== 'idle') {
						state.connectionStatus = 'error';
						state.connectionMessage = details.opened
							? '连接已断开，准备重连'
							: '推单握手失败，请检查 API 地址或 nginx WebSocket 代理';
						render();
						scheduleReconnect();
					}
				},
				onError: () => {
					state.connectionStatus = 'error';
					state.connectionMessage = '推单连接异常';
					render();
				},
				onMessage,
			});
		} catch (error) {
			state.connectionStatus = 'error';
			state.connectionMessage = error.message || '推单连接初始化失败';
			render();
			scheduleReconnect();
		}
	}

	function disconnectPOS(message) {
		clearReconnectTimer();
		destroySocket();
		reconnectAttempts = 0;
		state.connectionStatus = 'idle';
		state.connectionMessage = message || '已断开连接';
	}

	return {
		connectPOS,
		disconnectPOS,
	};
}
