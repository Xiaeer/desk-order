export function escapeHtml(value) {
	return String(value || '')
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');
}

export function formatAmount(value) {
	return `¥${(Number(value || 0) / 100).toFixed(2)}`;
}

export function formatTime(value) {
	const date = value ? new Date(value) : new Date();
	if (Number.isNaN(date.getTime())) {
		return '--';
	}
	const pad = (num) => String(num).padStart(2, '0');
	return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function formatCountdown(seconds) {
	const safeSeconds = Math.max(0, Number(seconds || 0));
	const mins = Math.floor(safeSeconds / 60);
	const secs = safeSeconds % 60;
	return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

export function formatOptionPriceDelta(value) {
	const delta = Number(value || 0);
	if (delta === 0) {
		return '';
	}
	const sign = delta > 0 ? '+' : '-';
	return `(${sign}${formatAmount(Math.abs(delta))})`;
}

export function getQRImageUrl(raw) {
	const value = String(raw || '').trim();
	if (!value) {
		return '';
	}
	return value.startsWith('data:image/') ? value : '';
}
