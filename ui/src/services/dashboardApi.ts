import { Transaction, TransactionStats } from '../types/dashboard'

async function parseJsonResponse<T>(response: Response, fallbackMessage: string): Promise<T> {
  if (!response.ok) {
    throw new Error(fallbackMessage)
  }

  return (await response.json()) as T
}

export async function fetchDashboardStats(): Promise<TransactionStats> {
  const response = await fetch('/api/stats')
  return parseJsonResponse<TransactionStats>(response, 'Failed to load dashboard stats')
}

export async function fetchRecentTransactions(limit = 100): Promise<Transaction[]> {
  const response = await fetch(`/api/transactions?limit=${limit}`)
  return parseJsonResponse<Transaction[]>(response, 'Failed to load recent transactions')
}
