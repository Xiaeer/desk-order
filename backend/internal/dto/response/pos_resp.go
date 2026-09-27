package response

type POSNativePayResp struct {
	OrderID     uint   `json:"order_id"`
	OrderNo     string `json:"order_no"`
	TotalAmount int    `json:"total_amount"`
	Status      int    `json:"status"`
	CodeURL     string `json:"code_url,omitempty"`
	Mode        string `json:"mode,omitempty"`
	MockResult  string `json:"mock_result,omitempty"`
}
