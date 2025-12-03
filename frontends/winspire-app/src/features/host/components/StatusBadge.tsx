/**
 * StatusBadge Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Displays tournament status with appropriate color coding
 */

import { memo } from 'react'
import { TOURNAMENT_STATUS_CONFIG } from '../constants'
import type { StatusBadgeProps } from '../types'

/**
 * Badge component for displaying tournament status
 * Colors: orange (scheduled), green (active), red (completed)
 */
export const StatusBadge = memo(function StatusBadge({ 
  status, 
  size = 'md' 
}: StatusBadgeProps) {
  const config = TOURNAMENT_STATUS_CONFIG[status]
  
  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-3 py-1 text-sm',
    lg: 'px-4 py-1.5 text-base'
  }

  return (
    <span
      className={`
        inline-flex items-center rounded-full font-medium border
        ${config.bgColor}
        ${config.textColor}
        ${config.borderColor}
        ${sizeClasses[size]}
      `}
    >
      {config.label}
    </span>
  )
})



