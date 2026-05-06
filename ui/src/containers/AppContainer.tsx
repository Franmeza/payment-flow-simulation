import App from '../App'
import { PROCESSING_FIXED_FEE, PROCESSING_PERCENTAGE_FEE } from '../constants/fees'
import { useDashboardData } from '../hooks/useDashboardData'

function AppContainer() {
  const dashboard = useDashboardData()

  return (
    <App
      transactions={dashboard.transactions}
      stats={dashboard.stats}
      loading={dashboard.loading}
      error={dashboard.error}
      lastUpdated={dashboard.lastUpdated}
      processingMarkupTotal={dashboard.processingMarkupTotal}
      estimatedNetVolume={dashboard.estimatedNetVolume}
      onRefresh={dashboard.refreshDashboard}
      processingPercentageFee={PROCESSING_PERCENTAGE_FEE}
      processingFixedFee={PROCESSING_FIXED_FEE}
    />
  )
}

export default AppContainer
