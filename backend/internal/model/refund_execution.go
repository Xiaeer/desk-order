package model

import "time"

const (
	RefundExecutionChannelWechat = "wechat"
	RefundExecutionChannelWallet = "wallet"

	RefundExecutionStatusPending    = "pending"
	RefundExecutionStatusProcessing = "processing"
	RefundExecutionStatusSuccess    = "success"
	RefundExecutionStatusFailed     = "failed"
)

type RefundExecution struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	RefundRequestID        uint       `gorm:"index" json:"refund_request_id"`
	ExecutionNo            string     `gorm:"size:64;uniqueIndex" json:"execution_no"`
	Channel                string     `gorm:"size:32;index" json:"channel"`
	OriginalPaySourceType  string     `gorm:"size:32;index" json:"original_pay_source_type"`
	OriginalPaySourceID    uint       `gorm:"index" json:"original_pay_source_id"`
	RechargeBatchID        *uint      `gorm:"index" json:"recharge_batch_id,omitempty"`
	ExecuteTotalAmount     int        `gorm:"default:0" json:"execute_total_amount"`
	ExecutePrincipalAmount int        `gorm:"default:0" json:"execute_principal_amount"`
	ExecuteGiftVoidAmount  int        `gorm:"default:0" json:"execute_gift_void_amount"`
	Status                 string     `gorm:"size:32;index" json:"status"`
	ExternalRefundNo       string     `gorm:"size:64;index" json:"external_refund_no"`
	ExternalTransactionID  string     `gorm:"size:64" json:"external_transaction_id"`
	ExternalResponseCode   string     `gorm:"size:64" json:"external_response_code"`
	ExternalResponseMsg    string     `gorm:"size:256" json:"external_response_msg"`
	RetryCount             int        `gorm:"default:0" json:"retry_count"`
	StartedAt              *time.Time `json:"started_at"`
	FinishedAt             *time.Time `json:"finished_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}
