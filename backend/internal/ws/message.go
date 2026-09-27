package ws

// 消息类型
const (
	MsgTypeNewOrder    = "new_order"    // 新订单推送
	MsgTypeOrderUpdate = "order_update" // 订单状态更新
	MsgTypePing        = "ping"
	MsgTypePong        = "pong"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
