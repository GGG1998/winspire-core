/**
 * useClipboard Hook
 * Shared hook for clipboard operations
 * 
 * Feature: 002-streamer-tournament-creation
 */

import { useState, useCallback, useRef, useEffect } from 'react'

interface UseClipboardReturn {
  copy: (text: string) => Promise<boolean>
  isCopied: boolean
  error: Error | null
  reset: () => void
}

/**
 * Duration to show "copied" feedback in milliseconds
 */
const COPY_FEEDBACK_DURATION = 2000

/**
 * Hook for copying text to clipboard with feedback state
 */
export function useClipboard(): UseClipboardReturn {
  const [isCopied, setIsCopied] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  /**
   * Copy text to clipboard
   * Returns true if successful, false otherwise
   */
  const copy = useCallback(async (text: string): Promise<boolean> => {
    // Clear previous timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    try {
      setError(null)

      // Try modern Clipboard API first
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text)
        setIsCopied(true)
        
        // Reset after duration
        timeoutRef.current = setTimeout(() => {
          setIsCopied(false)
        }, COPY_FEEDBACK_DURATION)
        
        return true
      }

      // Fallback for older browsers
      const success = fallbackCopy(text)
      if (success) {
        setIsCopied(true)
        
        timeoutRef.current = setTimeout(() => {
          setIsCopied(false)
        }, COPY_FEEDBACK_DURATION)
        
        return true
      }

      throw new Error('Clipboard API not supported')
    } catch (err) {
      const copyError = err instanceof Error ? err : new Error('Failed to copy')
      setError(copyError)
      setIsCopied(false)
      console.error('Failed to copy to clipboard:', err)
      return false
    }
  }, [])

  /**
   * Reset copied state manually
   */
  const reset = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    setIsCopied(false)
    setError(null)
  }, [])

  return {
    copy,
    isCopied,
    error,
    reset
  }
}

/**
 * Fallback copy method using execCommand
 * For browsers that don't support Clipboard API
 */
function fallbackCopy(text: string): boolean {
  const textarea = document.createElement('textarea')
  textarea.value = text
  
  // Avoid scrolling to bottom
  textarea.style.position = 'fixed'
  textarea.style.top = '0'
  textarea.style.left = '0'
  textarea.style.width = '2em'
  textarea.style.height = '2em'
  textarea.style.padding = '0'
  textarea.style.border = 'none'
  textarea.style.outline = 'none'
  textarea.style.boxShadow = 'none'
  textarea.style.background = 'transparent'
  textarea.style.opacity = '0'
  
  document.body.appendChild(textarea)
  textarea.select()
  
  let success = false
  try {
    success = document.execCommand('copy')
  } catch {
    success = false
  }
  
  document.body.removeChild(textarea)
  return success
}


