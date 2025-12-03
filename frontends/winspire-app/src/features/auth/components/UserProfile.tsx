import { useAuth } from '../hooks/useAuth';
import { Field, FieldGroup, Label } from '../../../shared/components/ui/fieldset';
import { Input } from '../../../shared/components/ui/input';
import { Heading, Subheading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';
import { Badge } from '../../../shared/components/ui/badge';
import { Divider } from '../../../shared/components/ui/divider';

export function UserProfile() {
  const { user } = useAuth();

  if (!user) {
    return (
      <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-700/50 dark:bg-zinc-900">
        <Text className="text-zinc-500 dark:text-zinc-400">
          No user data available. Please log in.
        </Text>
      </div>
    );
  }

  const profile = user.profile;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <Heading>Profile</Heading>
          <Text className="mt-2 text-zinc-500 dark:text-zinc-400">
            View your account information and profile details
          </Text>
        </div>
        <Badge color="zinc">
          {user.profileType === 'streamer' ? 'Streamer' : 'Player'}
        </Badge>
      </div>

      <Divider />

      {/* Profile Information */}
      <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-700/50 dark:bg-zinc-900">
        <Subheading className="mb-6">Profile Information</Subheading>
        
        <FieldGroup>
          <Field>
            <Label>Nickname</Label>
            <Input
              value={profile.nickname || 'Not set'}
              disabled
              className="bg-zinc-50 dark:bg-zinc-800"
            />
          </Field>

          <div className="grid gap-4 sm:grid-cols-2">
            <Field>
              <Label>First Name</Label>
              <Input
                value={profile.first_name || 'Not set'}
                disabled
                className="bg-zinc-50 dark:bg-zinc-800"
              />
            </Field>

            <Field>
              <Label>Last Name</Label>
              <Input
                value={profile.last_name || 'Not set'}
                disabled
                className="bg-zinc-50 dark:bg-zinc-800"
              />
            </Field>
          </div>

          <Field>
            <Label>Email Address</Label>
            <Input
              value={user.email}
              disabled
              className="bg-zinc-50 dark:bg-zinc-800"
            />
          </Field>

          <Divider className="my-6" />

          <div className="grid gap-4 sm:grid-cols-2">
            <Field>
              <Label>Country</Label>
              <Input
                value={profile.country_id || 'Not specified'}
                disabled
                className="bg-zinc-50 dark:bg-zinc-800"
              />
            </Field>

            <Field>
              <Label>City</Label>
              <Input
                value={profile.city || 'Not specified'}
                disabled
                className="bg-zinc-50 dark:bg-zinc-800"
              />
            </Field>
          </div>
        </FieldGroup>
      </div>

      {/* Account Details */}
      <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-700/50 dark:bg-zinc-900">
        <Subheading className="mb-6">Account Details</Subheading>
        
        <FieldGroup>
          <Field>
            <Label>User ID</Label>
            <Input
              value={user.id}
              disabled
              className="bg-zinc-50 font-mono text-sm dark:bg-zinc-800"
            />
          </Field>

          <div className="grid gap-4 sm:grid-cols-2">
            <Field>
              <Label>Account Created</Label>
              <Input
                value={profile.created_at ? new Date(profile.created_at).toLocaleDateString('pl-PL', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric'
                }) : 'Unknown'}
                disabled
                className="bg-zinc-50 dark:bg-zinc-800"
              />
            </Field>

            <Field>
              <Label>Last Updated</Label>
              <Input
                value={profile.updated_at ? new Date(profile.updated_at).toLocaleDateString('pl-PL', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric'
                }) : 'Unknown'}
                disabled
                className="bg-zinc-50 dark:bg-zinc-800"
              />
            </Field>
          </div>
        </FieldGroup>
      </div>

      {/* Info Banner */}
      <div className="rounded-lg bg-blue-50 p-4 ring-1 ring-blue-200 dark:bg-blue-900/20 dark:ring-blue-800">
        <Text className="text-sm text-blue-800 dark:text-blue-200">
          <strong>Note:</strong> Profile editing functionality will be available in a future update.
          If you need to update your information, please contact support.
        </Text>
      </div>
    </div>
  );
}
