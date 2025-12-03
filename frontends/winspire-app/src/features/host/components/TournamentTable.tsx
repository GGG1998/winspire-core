/**
 * TournamentTable Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Displays tournaments in a sortable table using TanStack Table
 */

import { useMemo } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  type ColumnDef,
  type SortingState
} from '@tanstack/react-table'
import { useState } from 'react'
import { StatusBadge } from './StatusBadge'
import { getStatusRowColor } from '../hooks/useTournamentStatus'
import { UI_LABELS } from '../constants'
import type { Tournament, TournamentTableProps } from '../types'

/**
 * Format date for display in Polish locale
 */
function formatDate(date: Date): string {
  // Check if date is valid
  if (!date || isNaN(date.getTime())) {
    return 'Nie ustalono'
  }
  
  return new Intl.DateTimeFormat('pl-PL', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

/**
 * Table component for displaying tournament list
 */
export function TournamentTable({ 
  tournaments, 
  isLoading,
  onRefresh,
  onTournamentClick
}: TournamentTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])

  // Note: Backend already filters draft tournaments for non-owners,
  // but we keep this as defense-in-depth
  // Tournaments list should only contain visible tournaments from API

  // Define table columns
  const columns = useMemo<ColumnDef<Tournament>[]>(() => [
    {
      accessorKey: 'name',
      header: UI_LABELS.table.columns.name,
      cell: (info) => (
        <span className="font-medium text-gray-900">
          {info.getValue() as string}
        </span>
      ),
      enableSorting: true
    },
    {
      accessorKey: 'startTime',
      header: UI_LABELS.table.columns.startTime,
      cell: (info) => formatDate(info.getValue() as Date),
      sortingFn: 'datetime',
      enableSorting: true
    },
    {
      accessorKey: 'game',
      header: UI_LABELS.table.columns.game,
      enableSorting: false
    },
    {
      id: 'status',
      header: UI_LABELS.table.columns.status,
      cell: ({ row }) => {
        return <StatusBadge status={row.original.status} size="sm" />
      },
      enableSorting: false
    }
  ], [])

  // Initialize table
  const table = useReactTable({
    data: tournaments,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel()
  })

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
        <span className="ml-3 text-gray-600">{UI_LABELS.loading}</span>
      </div>
    )
  }

  // Empty state
  if (tournaments.length === 0) {
    return (
      <div className="text-center py-12 px-4">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
          />
        </svg>
        <h3 className="mt-4 text-lg font-medium text-gray-900">
          Brak turniejów
        </h3>
        <p className="mt-2 text-gray-500 max-w-md mx-auto">
          {UI_LABELS.table.empty}
        </p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-gray-200">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th
                  key={header.id}
                  scope="col"
                  className={`
                    px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider
                    ${header.column.getCanSort() ? 'cursor-pointer hover:bg-gray-100 select-none' : ''}
                  `}
                  onClick={header.column.getToggleSortingHandler()}
                >
                  <div className="flex items-center gap-2">
                    {flexRender(
                      header.column.columnDef.header,
                      header.getContext()
                    )}
                    {header.column.getIsSorted() && (
                      <span className="text-blue-600">
                        {header.column.getIsSorted() === 'asc' ? '↑' : '↓'}
                      </span>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {table.getRowModel().rows.map((row) => {
            const rowColorClass = getStatusRowColor(row.original.status)

            return (
              <tr 
                key={row.id} 
                className={`${rowColorClass} hover:opacity-80 transition-opacity ${onTournamentClick ? 'cursor-pointer' : ''}`}
                onClick={() => onTournamentClick?.(row.original)}
              >
                {row.getVisibleCells().map((cell) => (
                  <td 
                    key={cell.id} 
                    className="px-6 py-4 whitespace-nowrap text-sm text-gray-700"
                  >
                    {flexRender(
                      cell.column.columnDef.cell,
                      cell.getContext()
                    )}
                  </td>
                ))}
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}


