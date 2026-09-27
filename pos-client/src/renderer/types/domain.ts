export type ConnectionStatus = 'idle' | 'connecting' | 'connected' | 'error';
export type ScreenName = 'orders' | 'cashier' | 'setup';
export type DocumentType = 'customer_receipt' | 'kitchen_ticket';

export interface POSConfig {
  apiBase: string;
  shopId: string;
  posToken: string;
  terminalName: string;
  customerPrinterName: string;
  kitchenPrinterName: string;
  autoPrintCustomer: boolean;
  autoPrintKitchen: boolean;
}

export interface OptionValue {
  id: string | number;
  name: string;
  price_delta?: number;
}

export interface OptionGroup {
  id: string | number;
  name: string;
  required?: boolean;
  select_type?: 'single' | 'multi' | string;
  min_select?: number;
  max_select?: number;
  values?: OptionValue[];
}

export interface Product {
  id: string | number;
  name: string;
  description?: string;
  price?: number;
  options?: OptionGroup[];
}

export interface MenuCategory {
  id?: string | number;
  name?: string;
  products?: Product[];
}

export interface MenuData {
  shop_name?: string;
  categories?: MenuCategory[];
}

export interface OrderItem {
  id?: number;
  name: string;
  quantity: number;
  price?: number;
  subtotal: number;
  option_summary?: string;
}

export interface OrderRecord {
  id: number;
  order_no: string;
  total_amount: number;
  remark?: string;
  status: number;
  created_at: string;
  table_no_snapshot?: string;
  shop_name?: string;
  items: OrderItem[];
}

export interface PaySessionRecord {
  payingOrder: Record<string, unknown> | null;
  payExpireAt: number;
  payCodeUrl: string;
  payQRCodeDataUrl: string;
  cashierMessage: string;
  pendingCreateRequestId: string;
}

export interface CartItem {
  cart_key: string;
  product_id: number;
  name: string;
  price: number;
  quantity: number;
  selected_options: Array<{ group_id: string; value_ids: string[] }>;
  option_summary: string;
}

export interface POSAuthPayload {
  shop_id: number;
  pos_token: string;
}

export interface POSWSMessage {
  type?: string;
  data?: unknown;
}
