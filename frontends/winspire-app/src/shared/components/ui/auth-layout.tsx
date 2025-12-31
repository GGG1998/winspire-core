import type React from 'react'

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex min-h-dvh flex-col p-2 bg-zinc-950 lg:bg-zinc-100 dark:bg-zinc-950">
      <div className="flex grow items-center justify-center p-6 bg-zinc-900 lg:rounded-lg lg:bg-white lg:p-10 lg:shadow-xs lg:ring-1 lg:ring-zinc-950/5 dark:bg-zinc-900 dark:lg:bg-zinc-900 dark:lg:ring-white/10">
        <div className="w-full max-w-sm">
          {children}
        </div>
      </div>
    </main>
  )
}
