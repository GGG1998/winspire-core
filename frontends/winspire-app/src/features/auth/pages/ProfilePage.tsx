import { UserProfile } from '../components/UserProfile';
import { AppLayout } from '../../../shared/components/layout/AppLayout';

export function ProfilePage() {
  return (
    <AppLayout>
      <div className="container mx-auto px-4 py-8">
        <UserProfile />
      </div>
    </AppLayout>
  );
}



