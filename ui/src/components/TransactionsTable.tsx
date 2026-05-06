import { useEffect, useMemo, useState } from 'react'
import { Transaction } from '../types/dashboard'

interface TransactionsTableProps {
  transactions: Transaction[]
  processingPercentageFee: number
  processingFixedFee: number
}

const amountFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
})

const PAGE_SIZE = 10

function TransactionsTable({
  transactions,
  processingPercentageFee,
  processingFixedFee,
}: TransactionsTableProps) {
  const [currentPage, setCurrentPage] = useState(1)

  const totalPages = Math.max(1, Math.ceil(transactions.length / PAGE_SIZE))

  useEffect(() => {
    if (currentPage > totalPages) {
      setCurrentPage(totalPages)
    }
  }, [currentPage, totalPages])

  const paginatedTransactions = useMemo(() => {
    const startIndex = (currentPage - 1) * PAGE_SIZE
    const endIndex = startIndex + PAGE_SIZE
    return transactions.slice(startIndex, endIndex)
  }, [currentPage, transactions])

  const pageStart = transactions.length === 0 ? 0 : (currentPage - 1) * PAGE_SIZE + 1
  const pageEnd = Math.min(currentPage * PAGE_SIZE, transactions.length)

  return (
    <section className="card transactions-card">
      <div className="card-header">
        <h2>Recent Merchant Transactions</h2>
        <span className="badge badge-light">
          {pageStart}-{pageEnd} of {transactions.length}
        </span>
      </div>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Transaction</th>
              <th>Merchant</th>
              <th>Amount</th>
              <th>Fee</th>
              <th>Est. Net</th>
              <th>Status</th>
              <th>Reason</th>
            </tr>
          </thead>
          <tbody>
            {paginatedTransactions.map((tx) => {
              const processingMarkup = tx.approved
                ? tx.amount * processingPercentageFee + processingFixedFee
                : 0
              const estimatedNet = tx.approved ? tx.amount - processingMarkup : 0

              return (
                <tr key={tx.id}>
                  <td>{new Date(tx.timestamp).toLocaleString()}</td>
                  <td className="mono">{tx.id}</td>
                  <td>{tx.merchant_id}</td>
                  <td>{amountFormatter.format(tx.amount)}</td>
                  <td>{tx.approved ? amountFormatter.format(processingMarkup) : '-'}</td>
                  <td>{tx.approved ? amountFormatter.format(estimatedNet) : '-'}</td>
                  <td>
                    <span className={tx.approved ? 'status success' : 'status error'}>
                      {tx.approved ? 'Approved' : 'Declined'}
                    </span>
                  </td>
                  <td>{tx.reason || '-'}</td>
                </tr>
              )
            })}
            {paginatedTransactions.length === 0 && (
              <tr>
                <td colSpan={8} className="empty-row">
                  No transactions yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="pagination-bar">
        <p className="pagination-meta">
          Page {currentPage} of {totalPages}
        </p>
        <div className="pagination-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => setCurrentPage((page) => Math.max(1, page - 1))}
            disabled={currentPage === 1}
          >
            Previous
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => setCurrentPage((page) => Math.min(totalPages, page + 1))}
            disabled={currentPage === totalPages}
          >
            Next
          </button>
        </div>
      </div>
    </section>
  )
}

export default TransactionsTable
