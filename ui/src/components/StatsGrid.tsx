import { TransactionStats } from '../types/dashboard'
import { formatCurrency } from '../lib/utils/formatter'

interface StatsGridProps {
  stats: TransactionStats | null
  processingMarkupTotal: number
  estimatedNetVolume: number
}

function StatsGrid({ stats, processingMarkupTotal, estimatedNetVolume }: StatsGridProps) {
  if (!stats) {
    return (
      <section className="stats-grid">
        <article className="stat-card">Loading stats...</article>
      </section>
    )
  }

  const approvalRate = `${(stats.approval_rate * 100).toFixed(1)}%`

  return (
    <section className="stats-grid">
      <article className="stat-card">
        <p className="stat-label">Total Transactions</p>
        <p className="stat-value">{stats.total_transactions}</p>
      </article>
      <article className="stat-card">
        <p className="stat-label">Approval Rate</p>
        <p className="stat-value">{approvalRate}</p>
      </article>
      <article className="stat-card">
        <p className="stat-label">Approved / Declined</p>
        <p className="stat-value">
          {stats.approved_count} / {stats.declined_count}
        </p>
      </article>
      <article className="stat-card">
        <p className="stat-label">Gross Volume</p>
        <p className="stat-value">{formatCurrency(stats.approved_volume)}</p>
      </article>
      <article className="stat-card">
        <p className="stat-label">Est. Fee Total (1.68% + $0.08)</p>
        <p className="stat-value">{formatCurrency(processingMarkupTotal)}</p>
      </article>
      <article className="stat-card">
        <p className="stat-label">Est. Net After Markup</p>
        <p className="stat-value">{formatCurrency(estimatedNetVolume)}</p>
      </article>
    </section>
  )
}

export default StatsGrid
