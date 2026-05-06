export interface Transaction {
  id: string
  card_uid: string
  merchant_id: string
  amount: number
  approved: boolean
  reason?: string
  timestamp: string
}

export interface Card {
  uid: string
  card_holder: string
  balance: number
  status: string
}

export interface TransactionStats {
  total_transactions: number
  approved_count: number
  declined_count: number
  approval_rate: number
  total_volume: number
  approved_volume: number
  declined_volume: number
}
