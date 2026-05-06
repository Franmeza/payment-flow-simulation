import { useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { PROCESSING_FIXED_FEE, PROCESSING_PERCENTAGE_FEE } from '../constants/fees'
import { fetchDashboardStats, fetchRecentTransactions } from '../services/dashboardApi'
import { Transaction, TransactionStats } from '../types/dashboard'

export interface DashboardViewModel {
  transactions: Transaction[]
  stats: TransactionStats | null
  loading: boolean
  error: string
  lastUpdated: string
  processingMarkupTotal: number
  estimatedNetVolume: number
  refreshDashboard: () => Promise<void>
}

const DASHBOARD_REFRESH_INTERVAL_MS = 5000
const TRANSACTIONS_LIMIT = 100

export function useDashboardData(): DashboardViewModel {
  const statsQuery = useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: fetchDashboardStats,
    refetchInterval: DASHBOARD_REFRESH_INTERVAL_MS,
  })

  const transactionsQuery = useQuery({
    queryKey: ['dashboard', 'transactions', TRANSACTIONS_LIMIT],
    queryFn: () => fetchRecentTransactions(TRANSACTIONS_LIMIT),
    refetchInterval: DASHBOARD_REFRESH_INTERVAL_MS,
  })

  const refreshDashboard = useCallback(async () => {
    await Promise.all([statsQuery.refetch(), transactionsQuery.refetch()])
  }, [statsQuery, transactionsQuery])

  const transactions = transactionsQuery.data ?? []
  const stats = statsQuery.data ?? null
  const loading = statsQuery.isPending || transactionsQuery.isPending

  const error = useMemo(() => {
    const queryError = statsQuery.error ?? transactionsQuery.error
    if (!queryError) {
      return ''
    }

    return queryError instanceof Error ? queryError.message : 'Dashboard request failed'
  }, [statsQuery.error, transactionsQuery.error])

  const lastUpdated = useMemo(() => {
    const latestUpdatedAt = Math.max(statsQuery.dataUpdatedAt, transactionsQuery.dataUpdatedAt)
    if (!latestUpdatedAt) {
      return ''
    }

    return new Date(latestUpdatedAt).toLocaleTimeString()
  }, [statsQuery.dataUpdatedAt, transactionsQuery.dataUpdatedAt])

  const { processingMarkupTotal, estimatedNetVolume } = useMemo(() => {
    const approvedTransactions = transactions.filter((tx) => tx.approved)
    const totalMarkup = approvedTransactions.reduce(
      (sum, tx) => sum + tx.amount * PROCESSING_PERCENTAGE_FEE + PROCESSING_FIXED_FEE,
      0,
    )
    const approvedGrossVolume = approvedTransactions.reduce((sum, tx) => sum + tx.amount, 0)

    return {
      processingMarkupTotal: totalMarkup,
      estimatedNetVolume: approvedGrossVolume - totalMarkup,
    }
  }, [transactions])

  return {
    transactions,
    stats,
    loading,
    error,
    lastUpdated,
    processingMarkupTotal,
    estimatedNetVolume,
    refreshDashboard,
  }
}
