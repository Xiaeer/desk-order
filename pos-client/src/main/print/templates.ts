type ReceiptItem = {
	name?: string;
	quantity?: number;
	subtotal?: number;
	option_summary?: string;
};

type ReceiptOrder = {
	order_no?: string;
	shop_name?: string;
	shop?: { name?: string };
	remark?: string;
	table_no_snapshot?: string;
	created_at?: string;
	paid_at?: string;
	total_amount?: number;
	items?: ReceiptItem[];
};

function escapeHtml(value: unknown): string {
	return String(value || '')
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/\"/g, '&quot;')
		.replace(/'/g, '&#39;');
}

function formatAmount(value: unknown): string {
	const amount = Number(value || 0) / 100;
	return amount.toFixed(2);
}

function formatDate(value: string | number | Date | null | undefined): string {
	const date = value ? new Date(value) : new Date();
	if (Number.isNaN(date.getTime())) {
		return '';
	}
	const pad = (num) => String(num).padStart(2, '0');
	return [
		date.getFullYear(),
		pad(date.getMonth() + 1),
		pad(date.getDate()),
	].join('-') + ' ' + [pad(date.getHours()), pad(date.getMinutes()), pad(date.getSeconds())].join(':');
}

function buildLineName(item: ReceiptItem): string {
	const optionSummary = item && item.option_summary ? `<div class="option">${escapeHtml(item.option_summary)}</div>` : '';
	return `<div class="item-name">${escapeHtml(item && item.name)}</div>${optionSummary}`;
}

function buildCustomerReceiptHtml(order: ReceiptOrder): string {
	const items = Array.isArray(order.items) ? order.items : [];
	const shopName = escapeHtml(order.shop_name || (order.shop && order.shop.name) || 'DeskOrder 门店');
	const remark = order.remark ? `<div class="remark">备注：${escapeHtml(order.remark)}</div>` : '';
	const tableNo = order.table_no_snapshot ? `<div>桌号：${escapeHtml(order.table_no_snapshot)}</div>` : '';
	const rows = items.length
		? items.map((item) => `
			<tr>
				<td class="name">${buildLineName(item)}</td>
				<td class="qty">x${escapeHtml(item.quantity)}</td>
				<td class="amount">${formatAmount(item.subtotal)}</td>
			</tr>
		`).join('')
		: '<tr><td class="name">订单明细未同步</td><td class="qty">-</td><td class="amount">-</td></tr>';

	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<title>DeskOrder Receipt</title>
	<style>
		@page {
			margin: 0;
		}
		body {
			margin: 0;
			padding: 10px;
			font-family: "Microsoft YaHei", "Segoe UI", sans-serif;
			color: #111111;
			background: #ffffff;
			font-size: 13px;
			line-height: 1.5;
			text-rendering: geometricPrecision;
			-webkit-print-color-adjust: exact;
			print-color-adjust: exact;
		}
		.receipt {
			width: 240px;
			margin: 0 auto;
		}
		.title {
			font-size: 20px;
			font-weight: 700;
			text-align: center;
			margin-bottom: 6px;
		}
		.sub {
			text-align: center;
			margin-bottom: 10px;
			font-size: 14px;
			font-weight: 700;
		}
		.meta {
			border-top: 1px dashed #333333;
			border-bottom: 1px dashed #333333;
			padding: 8px 0;
			margin-bottom: 8px;
		}
		.meta div,
		.remark {
			margin-bottom: 4px;
			word-break: break-all;
		}
		table {
			width: 100%;
			border-collapse: collapse;
			margin-bottom: 8px;
		}
		th,
		td {
			padding: 4px 0;
			vertical-align: top;
		}
		th {
			border-bottom: 1px dashed #333333;
			font-weight: 700;
			text-align: left;
		}
		.item-name {
			font-weight: 700;
		}
		.option {
			margin-top: 3px;
			font-size: 12px;
			color: #111111;
			word-break: break-all;
		}
		.qty,
		.amount {
			text-align: right;
			white-space: nowrap;
		}
		.total {
			border-top: 1px dashed #333333;
			padding-top: 8px;
			font-size: 14px;
			font-weight: 700;
			display: flex;
			justify-content: space-between;
		}
		.footer {
			margin-top: 10px;
			text-align: center;
			font-size: 12px;
			color: #111111;
		}
	</style>
</head>
<body>
	<div class="receipt">
		<div class="title">${shopName}</div>
		<div class="sub">顾客单</div>
		<div class="meta">
			<div>订单号：${escapeHtml(order.order_no || '')}</div>
			${tableNo}
			<div>下单时间：${escapeHtml(formatDate(order.created_at || order.paid_at))}</div>
			${remark}
		</div>
		<table>
			<thead>
				<tr>
					<th>商品</th>
					<th class="qty">数量</th>
					<th class="amount">金额</th>
				</tr>
			</thead>
			<tbody>${rows}</tbody>
		</table>
		<div class="total">
			<span>合计</span>
			<span>¥${formatAmount(order.total_amount)}</span>
		</div>
		<div class="footer">DeskOrder POS</div>
	</div>
</body>
</html>`;
}

function buildKitchenTicketHtml(order: ReceiptOrder): string {
	const items = Array.isArray(order.items) ? order.items : [];
	const shopName = escapeHtml(order.shop_name || (order.shop && order.shop.name) || 'DeskOrder 门店');
	const tableLine = order.table_no_snapshot
		? `<div class="headline strong">桌号 ${escapeHtml(order.table_no_snapshot)}</div>`
		: `<div class="headline">订单号 ${escapeHtml(order.order_no || '')}</div>`;
	const remark = order.remark ? `<div class="remark">备注：${escapeHtml(order.remark)}</div>` : '';
	const rows = items.length
		? items.map((item) => `
			<div class="item-row">
				<div class="item-left">${buildLineName(item)}</div>
				<div class="item-qty">x${escapeHtml(item.quantity)}</div>
			</div>
		`).join('')
		: '<div class="empty">订单明细未同步</div>';

	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8" />
	<title>DeskOrder Kitchen Ticket</title>
	<style>
		@page {
			margin: 0;
		}
		body {
			margin: 0;
			padding: 10px;
			font-family: "Microsoft YaHei", "Segoe UI", sans-serif;
			color: #111111;
			background: #ffffff;
			font-size: 14px;
			line-height: 1.5;
			text-rendering: geometricPrecision;
			-webkit-print-color-adjust: exact;
			print-color-adjust: exact;
		}
		.ticket {
			width: 240px;
			margin: 0 auto;
		}
		.title {
			font-size: 22px;
			font-weight: 800;
			text-align: center;
			margin-bottom: 4px;
		}
		.sub {
			text-align: center;
			font-size: 14px;
			margin-bottom: 10px;
		}
		.meta {
			border-top: 1px dashed #333333;
			border-bottom: 1px dashed #333333;
			padding: 8px 0;
			margin-bottom: 10px;
		}
		.headline {
			font-size: 14px;
			font-weight: 700;
			margin-bottom: 4px;
		}
		.strong {
			font-size: 22px;
			letter-spacing: 0.06em;
		}
		.meta-row,
		.remark {
			margin-bottom: 4px;
			word-break: break-all;
		}
		.items {
			display: grid;
			gap: 8px;
		}
		.item-row {
			display: flex;
			justify-content: space-between;
			gap: 12px;
			padding-bottom: 8px;
			border-bottom: 1px dashed #333333;
		}
		.item-left {
			flex: 1;
		}
		.item-name {
			font-size: 17px;
			font-weight: 800;
			word-break: break-all;
		}
		.option {
			margin-top: 4px;
			font-size: 13px;
			color: #111111;
			word-break: break-all;
		}
		.item-qty {
			font-size: 22px;
			font-weight: 800;
			white-space: nowrap;
		}
		.footer {
			margin-top: 10px;
			text-align: center;
			font-size: 12px;
			color: #111111;
		}
		.empty {
			padding: 8px 0;
		}
	</style>
</head>
<body>
	<div class="ticket">
		<div class="title">${shopName}</div>
		<div class="sub">后厨单</div>
		<div class="meta">
			${tableLine}
			<div class="meta-row">下单时间：${escapeHtml(formatDate(order.created_at || order.paid_at))}</div>
			<div class="meta-row">订单号：${escapeHtml(order.order_no || '')}</div>
			${remark}
		</div>
		<div class="items">${rows}</div>
		<div class="footer">DeskOrder Kitchen Ticket</div>
	</div>
</body>
</html>`;
}

export function buildReceiptHtml(order: ReceiptOrder, documentType?: string): string {
	if (documentType === 'kitchen_ticket') {
		return buildKitchenTicketHtml(order);
	}
	return buildCustomerReceiptHtml(order);
}
