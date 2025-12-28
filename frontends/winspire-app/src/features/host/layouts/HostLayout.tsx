/**
 * HostLayout Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Main layout wrapper for host/streamer pages with sidebar navigation
 */

import { useLocation } from 'react-router-dom'
import { Avatar } from '../../../shared/components/ui/avatar'
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner'
import {
  Dropdown,
  DropdownButton,
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '../../../shared/components/ui/dropdown'
import {
  Navbar,
  NavbarItem,
  NavbarSection,
  NavbarSpacer,
} from '../../../shared/components/ui/navbar'
import {
  Sidebar,
  SidebarBody,
  SidebarFooter,
  SidebarHeader,
  SidebarHeading,
  SidebarItem,
  SidebarLabel,
  SidebarSection,
  SidebarSpacer,
} from '../../../shared/components/ui/sidebar'
import { SidebarLayout } from '../../../shared/components/ui/sidebar-layout'
import { useAuth } from '../../auth'
import type { User } from '../../auth'
import {
  ArrowRightStartOnRectangleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  Cog8ToothIcon,
  LightBulbIcon,
  ShieldCheckIcon,
  UserIcon,
} from '@heroicons/react/16/solid'
import {
  Cog6ToothIcon,
  HomeIcon,
  InboxIcon,
  MagnifyingGlassIcon,
  QuestionMarkCircleIcon,
  SparklesIcon,
  TrophyIcon,
} from '@heroicons/react/20/solid'
import { useTournaments } from '../hooks/useTournaments'
import type { Tournament } from '../types'
import { useMemo } from 'react'

interface HostLayoutProps {
  children: React.ReactNode
}

export function HostLayout({ children }: HostLayoutProps) {
  const { user, isLoading, logout } = useAuth()

  if (isLoading && !user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-zinc-900">
        <LoadingSpinner />
      </div>
    )
  }

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-zinc-900 px-6">
        <div className="max-w-md text-center">
          <p className="text-lg font-semibold text-zinc-900 dark:text-white">Nie udało się załadować profilu.</p>
          <p className="mt-2 text-sm text-zinc-500 dark:text-zinc-400">Odśwież stronę lub zaloguj się ponownie.</p>
        </div>
      </div>
    )
  }

  return (
    <SidebarLayout
      navbar={<HostNavbar user={user} onLogout={logout} />}
      sidebar={<HostSidebar user={user} streamerId={user.id} onLogout={logout} />}
    >
      {children}
    </SidebarLayout>
  )
}

// ============================================================================
// HostSidebar
// ============================================================================

interface HostSidebarProps {
  user: User
  streamerId?: string
  onLogout: () => Promise<void>
}

function HostSidebar({ user, streamerId, onLogout }: HostSidebarProps) {
  const location = useLocation()
  const { tournaments } = useTournaments()

  // Navigation items
  const navigation = [
    { key: 'dashboard', label: 'Panel główny', href: `/h/${streamerId}/dashboard`, icon: HomeIcon },
    { key: 'tournaments', label: 'Turnieje', href: `/h/${streamerId}/tournaments`, icon: TrophyIcon },
    { key: 'settings', label: 'Ustawienia', href: `/h/${streamerId}/settings`, icon: Cog6ToothIcon },
  ]

  // Support links
  const supportLinks = [
    { key: 'support', label: 'Pomoc i wsparcie', href: '#', icon: QuestionMarkCircleIcon },
    { key: 'updates', label: 'Co nowego', href: '#', icon: SparklesIcon },
  ]

  // Determine active navigation key
  const getActiveNavKey = () => {
    if (location.pathname.includes('/tournaments')) return 'tournaments'
    if (location.pathname.includes('/settings')) return 'settings'
    if (location.pathname.includes('/dashboard')) return 'dashboard'
    return 'tournaments' // default
  }

  const activeNavKey = getActiveNavKey()
  const initials = getInitials(user)
  const streamerName = user.profile.nickname || `${user.profile.first_name} ${user.profile.last_name}`

  const handleLogout = () => {
    void onLogout()
  }

  // Get upcoming tournaments for sidebar
  const upcomingTournaments = tournaments
    .filter(t => !t.isCompleted && new Date(t.startTime) > new Date())
    .sort((a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime())
    .slice(0, 4)

  const activeMatch = useMemo(() => tournaments.find(t => t.me?.match?.matchId)?.me?.match, [tournaments])
  const returnMatchHref = activeMatch ? `/host/matches/${activeMatch.matchId}` : undefined

  return (
    <Sidebar>
      <SidebarHeader>
        <Dropdown>
          <DropdownButton as={SidebarItem} className="lg:mb-2.5">
            <Avatar initials={initials} />
            <SidebarLabel>{streamerName}</SidebarLabel>
            <ChevronDownIcon />
          </DropdownButton>
          <DropdownMenu className="min-w-80 lg:min-w-64" anchor="bottom start">
            <DropdownItem href="/auth/profile">
              <Cog8ToothIcon />
              <DropdownLabel>Ustawienia konta</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem href="/auth/profile">
              <Avatar slot="icon" initials={initials} />
              <DropdownLabel>{streamerName}</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={handleLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Wyloguj</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
        <SidebarSection className="max-lg:hidden">
          <SidebarItem href={`/h/${streamerId}/tournaments`}>
            <MagnifyingGlassIcon />
            <SidebarLabel>Szukaj turniejów</SidebarLabel>
          </SidebarItem>
        </SidebarSection>
      </SidebarHeader>

      <SidebarBody>
        <SidebarSection>
          {navigation.map(item => (
            <SidebarItem key={item.key} href={item.href} current={activeNavKey === item.key}>
              <item.icon />
              <SidebarLabel>{item.label}</SidebarLabel>
            </SidebarItem>
          ))}
        </SidebarSection>

        {upcomingTournaments.length > 0 && (
          <SidebarSection className="max-lg:hidden">
            <SidebarHeading>Nadchodzące turnieje</SidebarHeading>
            {upcomingTournaments.map(tournament => (
              <SidebarItem key={tournament.id} href={`/h/${streamerId}/tournaments/${tournament.id}`}>
                {formatTournamentLabel(tournament)}
              </SidebarItem>
            ))}
          </SidebarSection>
        )}

        <SidebarSpacer />

        <SidebarSection>
          {returnMatchHref && (
            <SidebarItem href={returnMatchHref}>
              <SparklesIcon />
              <SidebarLabel>Wróć do meczu</SidebarLabel>
            </SidebarItem>
          )}
          {supportLinks.map(link => (
            <SidebarItem key={link.key} href={link.href}>
              <link.icon />
              <SidebarLabel>{link.label}</SidebarLabel>
            </SidebarItem>
          ))}
        </SidebarSection>
      </SidebarBody>

      <SidebarFooter className="max-lg:hidden">
        <Dropdown>
          <DropdownButton as={SidebarItem}>
            <span className="flex min-w-0 items-center gap-3">
              <Avatar className="size-10" initials={initials} square alt="" />
              <span className="min-w-0">
                <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">
                  {streamerName}
                </span>
                <span className="block truncate text-xs/5 font-normal text-zinc-500 dark:text-zinc-400">
                  {user.email}
                </span>
              </span>
            </span>
            <ChevronUpIcon />
          </DropdownButton>
          <DropdownMenu className="min-w-64" anchor="top start">
            <DropdownItem href="/auth/profile">
              <UserIcon />
              <DropdownLabel>Zobacz profil</DropdownLabel>
            </DropdownItem>
            <DropdownItem href="/auth/profile">
              <ShieldCheckIcon />
              <DropdownLabel>Bezpieczeństwo</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem href="#">
              <LightBulbIcon />
              <DropdownLabel>Prześlij opinię</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={handleLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Wyloguj</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </SidebarFooter>
    </Sidebar>
  )
}

// ============================================================================
// HostNavbar
// ============================================================================

interface HostNavbarProps {
  user: User
  onLogout: () => Promise<void>
}

function HostNavbar({ user, onLogout }: HostNavbarProps) {
  const initials = getInitials(user)

  const handleLogout = () => {
    void onLogout()
  }

  return (
    <Navbar>
      <NavbarSpacer />
      <NavbarSection>
        <NavbarItem aria-label="Szukaj">
          <MagnifyingGlassIcon />
        </NavbarItem>
        <NavbarItem aria-label="Powiadomienia">
          <InboxIcon />
        </NavbarItem>
        <Dropdown>
          <DropdownButton as={NavbarItem}>
            <Avatar initials={initials} square />
          </DropdownButton>
          <DropdownMenu className="min-w-64" anchor="bottom end">
            <DropdownItem href="/auth/profile">
              <UserIcon />
              <DropdownLabel>Profil</DropdownLabel>
            </DropdownItem>
            <DropdownItem href="/auth/profile">
              <Cog8ToothIcon />
              <DropdownLabel>Ustawienia</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={handleLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Wyloguj</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </NavbarSection>
    </Navbar>
  )
}

// ============================================================================
// Helper Functions
// ============================================================================

function getInitials(user: User): string {
  const first = user.profile.first_name?.[0] ?? ''
  const last = user.profile.last_name?.[0] ?? ''
  const nickname = user.profile.nickname?.slice(0, 2) ?? ''
  return nickname || `${first}${last}` || user.email.slice(0, 2)
}

function formatTournamentLabel(tournament: Tournament): string {
  const date = new Date(tournament.startTime)
  const day = date.getDate()
  const month = date.toLocaleDateString('pl-PL', { month: 'short' })
  const time = date.toLocaleTimeString('pl-PL', { hour: '2-digit', minute: '2-digit' })
  
  // Truncate name if too long
  const name = tournament.name.length > 20 
    ? tournament.name.slice(0, 17) + '...' 
    : tournament.name
  
  return `${name} · ${day} ${month} ${time}`
}

