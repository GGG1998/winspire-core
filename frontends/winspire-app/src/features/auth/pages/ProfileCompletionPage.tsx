import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { ProfileCompletionForm } from '../components/ProfileCompletionForm';
import { AuthLayout } from '../../../shared/components/ui/auth-layout';
import { Heading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';

export function ProfileCompletionPage() {
  const { user, isLoading } = useAuth();
  const navigate = useNavigate();
  const [initialData, setInitialData] = useState<{
    nickname?: string;
    // first_name?: string;
    // last_name?: string;
  }>({});

  useEffect(() => {
    if (!isLoading && user) {
      // Check if profile is already complete
      const profile = user.profile;
      if (profile.nickname) {
        // Profile is complete, redirect to home
        navigate('/');
        return;
      }

      // Set initial data from profile
      setInitialData({
        nickname: profile.nickname || '',
        // first_name: profile.first_name || '',
        // last_name: profile.last_name || '',
      });
    }
  }, [user, isLoading, navigate]);

  if (isLoading) {
    return (
      <AuthLayout>
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      </AuthLayout>
    );
  }

  if (!user) {
    // No user, redirect to login
    navigate('/auth/login');
    return null;
  }

  return (
    <AuthLayout>
      <div className="space-y-6">
        <div className="text-center space-y-2">
          <Heading>Complete Your Profile</Heading>
          <Text className="text-zinc-500 dark:text-zinc-400">
            Please provide the following information to continue using the platform.
          </Text>
          <div className="rounded-lg bg-amber-50 p-4 ring-1 ring-amber-200 dark:bg-amber-900/20 dark:ring-amber-800">
            <Text className="text-sm text-amber-800 dark:text-amber-200">
              ⚠️ You must complete your profile to access tournaments and other features.
            </Text>
          </div>
        </div>

        <ProfileCompletionForm
          userId={user.id}
          profileType={user.profileType}
          initialData={initialData}
        />
      </div>
    </AuthLayout>
  );
}




