import { Transaction } from '../types/dashboard'
import { formatCurrency } from '../lib/utils/formatter'

interface MobileTransactionsListProps {
  transactions: Transaction[]
  processingPercentageFee: number
  processingFixedFee: number
}

function MobileTransactionsList({
  transactions,
  processingPercentageFee,
  processingFixedFee,
}: MobileTransactionsListProps) {
  return (
    <div className="mobile-transactions-list">
      {transactions.map((tx) => {
        const processingMarkup = tx.approved
          ? tx.amount * processingPercentageFee + processingFixedFee
          : 0
        const estimatedNet = tx.approved ? tx.amount - processingMarkup : 0

        return (
          <article key={`mobile-${tx.id}`} className="mobile-transaction-card">
            <div className="mobile-transaction-header">
              <p className="mobile-transaction-id mono">{tx.id}</p>
              <span className={tx.approved ? 'status success' : 'status error'}>
                {tx.approved ? 'Approved' : 'Declined'}
              </span>
            </div>
            <div className="mobile-transaction-grid">
              <p>
                <span>Timestamp</span>
                {new Date(tx.timestamp).toLocaleString()}
              </p>
              <p>
                <span>Merchant</span>
                {tx.merchant_id}
              </p>
              <p>
                <span>Amount</span>
                {formatCurrency(tx.amount)}
              </p>
              <p>
                <span>Fee</span>
                {tx.approved ? formatCurrency(processingMarkup) : '-'}
              </p>
              <p>
                <span>Est. Net</span>
                {tx.approved ? formatCurrency(estimatedNet) : '-'}
              </p>
              <p>
                <span>Reason</span>
                {tx.reason || '-'}
              </p>
            </div>
          </article>
        )
      })}
      {transactions.length === 0 && <div className="empty-row mobile-empty-row">No transactions yet.</div>}
    </div>
  )
}

export default MobileTransactionsList
