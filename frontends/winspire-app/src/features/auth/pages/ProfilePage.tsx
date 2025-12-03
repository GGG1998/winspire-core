import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';
import { UserProfile } from '../components/UserProfile';
import { SidebarLayout } from '../../../shared/components/ui/sidebar-layout';
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
} from '../../../shared/components/ui/sidebar';
import {
  Dropdown,
  DropdownButton,
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
} from '../../../shared/components/ui/dropdown';
import { Avatar } from '../../../shared/components/ui/avatar';
import {
  ArrowRightStartOnRectangleIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  Cog8ToothIcon,
  HomeIcon,
  ShieldCheckIcon,
  UserIcon,
} from '@heroicons/react/16/solid';
import { InboxIcon, MagnifyingGlassIcon } from '@heroicons/react/20/solid';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { TrophyIcon } from '@heroicons/react/20/solid';

export function ProfilePage() {
  const { user, isLoading, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/auth/login');
  };

  if (isLoading && !user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-white dark:bg-zinc-950">
        <LoadingSpinner />
      </div>
    );
  }

  if (!user) {
    navigate('/auth/login');
    return null;
  }

  const initials = getInitials(user);
  const displayName = user.profile.nickname || `${user.profile.first_name} ${user.profile.last_name}`;

  return (
    <SidebarLayout
      navbar={<ProfileNavbar user={user} onLogout={handleLogout} initials={initials} />}
      sidebar={<ProfileSidebar user={user} onLogout={handleLogout} initials={initials} displayName={displayName} />}
    >
      <div className="mx-auto max-w-4xl py-6">
        <UserProfile />
      </div>
    </SidebarLayout>
  );
}

function ProfileNavbar({ user, onLogout, initials }: { user: any; onLogout: () => void; initials: string }) {
  return (
    <Navbar>
      <NavbarSpacer />
      <NavbarSection>
        <NavbarItem aria-label="Search">
          <MagnifyingGlassIcon />
        </NavbarItem>
        <NavbarItem aria-label="Notifications">
          <InboxIcon />
        </NavbarItem>
        <Dropdown>
          <DropdownButton as={NavbarItem}>
            <Avatar initials={initials} square />
          </DropdownButton>
          <DropdownMenu className="min-w-64" anchor="bottom end">
            <DropdownItem href="/auth/profile">
              <UserIcon />
              <DropdownLabel>Profile</DropdownLabel>
            </DropdownItem>
            <DropdownItem href="/">
              <HomeIcon />
              <DropdownLabel>Home</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={onLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Sign Out</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </NavbarSection>
    </Navbar>
  );
}

function ProfileSidebar({
  user,
  onLogout,
  initials,
  displayName,
}: {
  user: any;
  onLogout: () => void;
  initials: string;
  displayName: string;
}) {
  const navigation = [
    { key: 'home', label: 'Home', href: '/', icon: HomeIcon },
    { key: 'profile', label: 'Profile Settings', href: '/auth/profile', icon: UserIcon },
    { key: 'my-tournaments', label: 'My Tournaments', href: `/h/${user.id}/tournaments`, icon: TrophyIcon },
  ];

  return (
    <Sidebar>
      <SidebarHeader>
        <Dropdown>
          <DropdownButton as={SidebarItem} className="lg:mb-2.5">
            <Avatar initials={"WS"} />
            <SidebarLabel>Winspire</SidebarLabel>
            {/* <ChevronDownIcon /> */}
          </DropdownButton>
          {/* <DropdownMenu className="min-w-80 lg:min-w-64" anchor="bottom start">
            <DropdownItem href="/auth/profile">
              <Cog8ToothIcon />
              <DropdownLabel>Account Settings</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem href="/auth/profile">
              <Avatar slot="icon" initials={initials} />
              <DropdownLabel>{displayName}</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={onLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Sign Out</DropdownLabel>
            </DropdownItem>
          </DropdownMenu> */}
        </Dropdown>
      </SidebarHeader>
      <SidebarBody>
        <SidebarSection>
          <SidebarHeading>Navigation</SidebarHeading>
          {navigation.map((item) => (
            <SidebarItem key={item.key} href={item.href} current={item.key === 'profile'}>
              <item.icon />
              <SidebarLabel>{item.label}</SidebarLabel>
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
              <DropdownLabel>View Profile</DropdownLabel>
            </DropdownItem>
            <DropdownItem href="/auth/profile">
              <ShieldCheckIcon />
              <DropdownLabel>Security</DropdownLabel>
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem onClick={onLogout}>
              <ArrowRightStartOnRectangleIcon />
              <DropdownLabel>Sign Out</DropdownLabel>
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
      </SidebarFooter>
    </Sidebar>
  );
}

function getInitials(user: any): string {
  const first = user.profile.first_name?.[0] ?? '';
  const last = user.profile.last_name?.[0] ?? '';
  const nickname = user.profile.nickname?.slice(0, 2) ?? '';
  return nickname || `${first}${last}` || user.email.slice(0, 2);
}
