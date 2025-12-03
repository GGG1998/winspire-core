/**
 * useCountdown Hook
 * Feature: 002-streamer-tournament-creation
 * 
 * Provides countdown timer functionality for tournament start times
 */

import { useState, useEffect, useMemo } from 'react'

export interface CountdownTime {
  days: number
  hours: number
  minutes: number
  seconds: number
  totalSeconds: number
  isExpired: boolean
}

export interface UseCountdownReturn {
  time: CountdownTime
  formatted: string
  formattedShort: string
  isExpired: boolean
}

/**
 * Calculate time remaining until target date
 */
function calculateTimeRemaining(targetDate: Date): CountdownTime {
  const now = new Date().getTime()
  const target = targetDate.getTime()
  const difference = target - now

  if (difference <= 0) {
    return {
      days: 0,
      hours: 0,
      minutes: 0,
      seconds: 0,
      totalSeconds: 0,
      isExpired: true
    }
  }

  const totalSeconds = Math.floor(difference / 1000)
  const days = Math.floor(totalSeconds / (60 * 60 * 24))
  const hours = Math.floor((totalSeconds % (60 * 60 * 24)) / (60 * 60))
  const minutes = Math.floor((totalSeconds % (60 * 60)) / 60)
  const seconds = totalSeconds % 60

  return {
    days,
    hours,
    minutes,
    seconds,
    totalSeconds,
    isExpired: false
  }
}

/**
 * Format countdown for display (e.g., "13:19:51")
 */
function formatCountdown(time: CountdownTime): string {
  if (time.isExpired) return '00:00:00'
  
  const parts: string[] = []
  
  if (time.days > 0) {
    parts.push(`${time.days}d`)
  }
  
  const hours = String(time.hours).padStart(2, '0')
  const minutes = String(time.minutes).padStart(2, '0')
  const seconds = String(time.seconds).padStart(2, '0')
  
  if (time.days > 0) {
    parts.push(`${hours}:${minutes}:${seconds}`)
    return parts.join(' ')
  }
  
  return `${hours}:${minutes}:${seconds}`
}

/**
 * Format countdown in human-readable form (e.g., "In about 13 hours")
 */
function formatCountdownHuman(time: CountdownTime): string {
  if (time.isExpired) return 'Started'
  
  if (time.days > 1) {
    return `In about ${time.days} days`
  }
  
  if (time.days === 1) {
    return 'In about 1 day'
  }
  
  if (time.hours > 1) {
    return `In about ${time.hours} hours`
  }
  
  if (time.hours === 1) {
    return 'In about 1 hour'
  }
  
  if (time.minutes > 1) {
    return `In about ${time.minutes} minutes`
  }
  
  if (time.minutes === 1) {
    return 'In about 1 minute'
  }
  
  return 'Starting soon'
}

/**
 * Hook for countdown timer to a target date
 */
export function useCountdown(targetDate: Date | undefined): UseCountdownReturn {
  const [time, setTime] = useState<CountdownTime>(() => 
    targetDate ? calculateTimeRemaining(targetDate) : {
      days: 0,
      hours: 0,
      minutes: 0,
      seconds: 0,
      totalSeconds: 0,
      isExpired: true
    }
  )

  useEffect(() => {
    if (!targetDate) return

    // Initial calculation
    setTime(calculateTimeRemaining(targetDate))

    // Update every second
    const interval = setInterval(() => {
      const newTime = calculateTimeRemaining(targetDate)
      setTime(newTime)
      
      // Clear interval if expired
      if (newTime.isExpired) {
        clearInterval(interval)
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [targetDate])

  const formatted = useMemo(() => formatCountdown(time), [time])
  const formattedShort = useMemo(() => formatCountdownHuman(time), [time])

  return {
    time,
    formatted,
    formattedShort,
    isExpired: time.isExpired
  }
}


