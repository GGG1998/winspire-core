/**
 * LobbyLayout Component
 * Feature: 003-matchmaking-lobby-frontend
 * 
 * Main layout wrapper for player lobby pages with sidebar navigation
 */

import { useLocation } from 'react-router-dom';
import { Avatar } from '../../../shared/components/ui/avatar';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import {
  Dropdown,
  DropdownButton,
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '../../../shared/components/ui/dropdown';
import {
  Navbar,
  NavbarItem,
  NavbarSection,
  NavbarSpacer,
} from '../../../shared/components/ui/navbar';
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
} from '../../../shared/components/ui/sidebar';
import { SidebarLayout } from '../../../shared/components/ui/sidebar-layout';
import { useAuth } from '../../auth';
import type { User } from '../../auth';
import {
  ArrowRightStartOnRectangleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  Cog8ToothIcon,
  UserIcon,
} from '@heroicons/react/16/solid';
import {
  HomeIcon,
  InboxIcon,
  MagnifyingGlassIcon,
  TrophyIcon,
  PlayIcon,
  ChartBarIcon,
  ArrowLeftIcon,
} from '@heroicons/react/20/solid';

interface LobbyLayoutProps {
  children: React.ReactNode;
  tournamentId?: string;
  streamerId?: string;
}

export function LobbyLayout({ children, tournamentId, streamerId }: LobbyLayoutProps) {
  const { user, isLoading, logout } = useAuth();

  if (isLoading && !user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-zinc-900">
        <LoadingSpinner />
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-zinc-900 px-6">
        <div className="max-w-md text-center">
          <p className="text-lg font-semibold text-zinc-900 dark:text-white">
            Nie udało się załadować profilu.
          </p>
          <p className="mt-2 text-sm text-zinc-500 dark:text-zinc-400">
            Odśwież stronę lub zaloguj się ponownie.
          </p>
        </div>
      </div>
    );
  }

  return (
    <SidebarLayout
      navbar={<LobbyNavbar user={user} onLogout={logout} />}
      sidebar={<LobbySidebar user={user} tournamentId={tournamentId} streamerId={streamerId} onLogout={logout} />}
    >
      {children}
    </SidebarLayout>
  );
}

// ============================================================================
// LobbySidebar
// ============================================================================

interface LobbySidebarProps {
  user: User;
  tournamentId?: string;
  streamerId?: string;
  onLogout: () => Promise<void>;
}

function LobbySidebar({ user, tournamentId, streamerId, onLogout }: LobbySidebarProps) {
  const location = useLocation();

  // Navigation items for players
  const navigation =[
    { 
      key: 'tournament-detail', 
      label: 'Powrót do Podglądu turnieju', 
      href: `/h/${streamerId}/tournaments/${tournamentId}`, 
      icon: ArrowLeftIcon 
    },
  ];

  // Determine active navigation key
  const getActiveNavKey = () => {
    if (tournamentId && streamerId) return 'tournament-detail';
    if (location.pathname === '/') return 'home';
    return 'tournaments'; // default
  };

  const activeNavKey = getActiveNavKey();
  const initials = getInitials(user);
  const displayName = user.profile.nickname || `${user.profile.first_name} ${user.profile.last_name}`;

  const handleLogout = () => {
    void onLogout();
  };

  return (
    <Sidebar>
      <SidebarHeader>
        <Dropdown>
          <DropdownButton as={SidebarItem} className="lg:mb-2.5">
            <Avatar initials={initials} />
            <SidebarLabel>{displayName}</SidebarLabel>
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
              <DropdownLabel>{displayName}</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={handleLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Wyloguj</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </SidebarHeader>

      <SidebarBody>
        <SidebarSection>
          {navigation.map((item) => (
            <SidebarItem key={item.key} href={item.href} current={activeNavKey === item.key}>
              <item.icon />
              <SidebarLabel>{item.label}</SidebarLabel>
            </SidebarItem>
          ))}
        </SidebarSection>

        <SidebarSpacer />

        <SidebarSection>
          <SidebarHeading>W turnieju</SidebarHeading>
          <div className="px-2 py-3">
            <p className="text-xs text-zinc-500 dark:text-zinc-400">
              Jesteś aktualnie w poczekalni turniejowej
            </p>
          </div>
        </SidebarSection>
      </SidebarBody>

      <SidebarFooter className="max-lg:hidden">
        <Dropdown>
          <DropdownButton as={SidebarItem}>
            <span className="flex min-w-0 items-center gap-3">
              <Avatar className="size-10" initials={initials} square alt="" />
              <span className="min-w-0">
                <span className="block truncate text-sm/5 font-medium text-zinc-950 dark:text-white">
                  {displayName}
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
            <DropdownDivider />
            <DropdownItem onClick={handleLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Wyloguj</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </SidebarFooter>
    </Sidebar>
  );
}

// ============================================================================
// LobbyNavbar
// ============================================================================

interface LobbyNavbarProps {
  user: User;
  onLogout: () => Promise<void>;
}

function LobbyNavbar({ user, onLogout }: LobbyNavbarProps) {
  const initials = getInitials(user);

  const handleLogout = () => {
    void onLogout();
  };

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
  );
}

// ============================================================================
// Helper Functions
// ============================================================================

function getInitials(user: User): string {
  const first = user.profile.first_name?.[0] ?? '';
  const last = user.profile.last_name?.[0] ?? '';
  const nickname = user.profile.nickname?.slice(0, 2) ?? '';
  return nickname || `${first}${last}` || user.email.slice(0, 2);
}

