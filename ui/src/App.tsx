import Hero from './components/Hero'
import StatsGrid from './components/StatsGrid'
import TransactionsTable from './components/TransactionsTable'
import { Transaction, TransactionStats } from './types/dashboard'

interface AppProps {
  transactions: Transaction[]
  stats: TransactionStats | null
  loading: boolean
  error: string
  lastUpdated: string
  processingMarkupTotal: number
  estimatedNetVolume: number
  onRefresh: () => Promise<void> | void
  processingPercentageFee: number
  processingFixedFee: number
}

function App({
  transactions,
  stats,
  loading,
  error,
  lastUpdated,
  processingMarkupTotal,
  estimatedNetVolume,
  onRefresh,
  processingPercentageFee,
  processingFixedFee,
}: AppProps) {

  return (
    <div className="app-shell">
      <Hero />
      <section className="toolbar">
        <button
          type="button"
          className="btn btn-primary"
          onClick={() => {
            void onRefresh()
          }}
        >
          Refresh Now
        </button>
        <p className="toolbar-meta">
          Polling every 5s{lastUpdated ? ` · Last update ${lastUpdated}` : ''}
        </p>
      </section>
      {error && <p className="error-banner">{error}</p>}
      <StatsGrid
        stats={stats}
        processingMarkupTotal={processingMarkupTotal}
        estimatedNetVolume={estimatedNetVolume}
      />

      <main className="content-grid">
        <TransactionsTable
          transactions={transactions}
          processingPercentageFee={processingPercentageFee}
          processingFixedFee={processingFixedFee}
        />
      </main>
      {loading && <p className="loading-note">Loading dashboard data...</p>}
    </div>
  )
}

export default App
