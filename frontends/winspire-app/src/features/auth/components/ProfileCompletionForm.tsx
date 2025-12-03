import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuth } from '../hooks/useAuth';
// import { useCountries } from '../hooks/useCountries';
import { Button } from '../../../shared/components/ui/button';
import { Input, InputGroup } from '../../../shared/components/ui/input';
// import { Select } from '../../../shared/components/ui/select';
import { Field, FieldGroup, Label, ErrorMessage as FieldErrorMessage } from '../../../shared/components/ui/fieldset';
import { Text } from '../../../shared/components/ui/text';
import { profileCompletionSchema, type ProfileCompletionInput } from '../schemas';
import { supabase } from '../../../shared/api/supabase';

interface ProfileCompletionFormProps {
  userId: string;
  profileType: 'user' | 'streamer';
  initialData?: {
    nickname?: string;
    // first_name?: string;
    // last_name?: string;
  };
}

export function ProfileCompletionForm({ userId, profileType, initialData }: ProfileCompletionFormProps) {
  const navigate = useNavigate();
  const { refreshUser } = useAuth();
  // const { countries, isLoading: countriesLoading } = useCountries();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ProfileCompletionInput>({
    resolver: zodResolver(profileCompletionSchema),
    defaultValues: {
      nickname: initialData?.nickname || '',
      // first_name: initialData?.first_name || '',
      // last_name: initialData?.last_name || '',
    },
  });

  const onSubmit = async (data: ProfileCompletionInput) => {
    console.log('[ProfileCompletion] onSubmit started', { data, userId, profileType });
    setError(null);
    setIsLoading(true);

    try {
      const tableName = profileType === 'streamer' ? 'streamer_profiles' : 'user_profiles';
      
      // Only send nickname (country_id and city not used for now)
      const updateData = {
        nickname: data.nickname,
      };

      console.log('[ProfileCompletion] Updating profile...', { tableName, updateData });
      const { error: updateError } = await supabase
        .from(tableName)
        .update(updateData)
        .eq('id', userId);

      if (updateError) {
        console.error('[ProfileCompletion] Update failed:', updateError);
        throw new Error(updateError.message);
      }
      console.log('[ProfileCompletion] ✓ Profile updated successfully');

      // MUST refresh user context BEFORE navigate to avoid race condition with ProtectedRoute
      // ProtectedRoute checks user.profile.nickname - if not refreshed, it redirects back here
      try {
        console.log('[ProfileCompletion] Refreshing user context (skipping session refresh)...');
        
        // Skip refreshSession() as it may hang with JWT hook issues
        // Just refresh user context directly from database
        await Promise.race([
          refreshUser(),
          new Promise((_, reject) => 
            setTimeout(() => reject(new Error('User context refresh timeout')), 3000)
          )
        ]);
        console.log('[ProfileCompletion] ✓ User context refreshed');
      } catch (refreshErr) {
        console.error('[ProfileCompletion] Refresh failed:', refreshErr);
        // Continue anyway - worst case user will re-login
      }

      console.log('[ProfileCompletion] Navigating to /');
      // Now safe to navigate - user context has updated nickname
      navigate('/');
      console.log('[ProfileCompletion] ✓ Navigate called');
    } catch (err) {
      console.error('[ProfileCompletion] Error in onSubmit:', err);
      setError(err instanceof Error ? err.message : 'Failed to update profile. Please try again.');
    } finally {
      console.log('[ProfileCompletion] onSubmit finished, setIsLoading(false)');
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <FieldGroup>
        {error && (
          <div className="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 dark:bg-red-900/20 dark:ring-red-800">
            <Text className="text-red-800 dark:text-red-200">{error}</Text>
          </div>
        )}

        <Field>
          <Label htmlFor="nickname">Nickname *</Label>
          <InputGroup>
            <Input
              id="nickname"
              type="text"
              {...register('nickname')}
              data-invalid={errors.nickname ? '' : undefined}
              placeholder="Your gaming nickname"
            />
          </InputGroup>
          {errors.nickname && (
            <FieldErrorMessage>{errors.nickname.message}</FieldErrorMessage>
          )}
        </Field>

        {/* <Field>
          <Label htmlFor="first_name">First Name *</Label>
          <InputGroup>
            <Input
              id="first_name"
              type="text"
              {...register('first_name')}
              data-invalid={errors.first_name ? '' : undefined}
            />
          </InputGroup>
          {errors.first_name && (
            <FieldErrorMessage>{errors.first_name.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="last_name">Last Name *</Label>
          <InputGroup>
            <Input
              id="last_name"
              type="text"
              {...register('last_name')}
              data-invalid={errors.last_name ? '' : undefined}
            />
          </InputGroup>
          {errors.last_name && (
            <FieldErrorMessage>{errors.last_name.message}</FieldErrorMessage>
          )}
        </Field> */}

        {/* Country and City fields disabled for now
        <Field>
          <Label htmlFor="country_id">Country (Optional)</Label>
          <Select
            id="country_id"
            {...register('country_id')}
            disabled={countriesLoading}
            data-invalid={errors.country_id ? '' : undefined}
          >
            <option value="">Select a country</option>
            {countries.map((country) => (
              <option key={country.id} value={country.id}>
                {country.name}
              </option>
            ))}
          </Select>
          {errors.country_id && (
            <FieldErrorMessage>{errors.country_id.message}</FieldErrorMessage>
          )}
        </Field>
        
        <Field>
          <Label htmlFor="city">City (Optional)</Label>
          <InputGroup>
            <Input
              id="city"
              type="text"
              {...register('city')}
              data-invalid={errors.city ? '' : undefined}
            />
          </InputGroup>
          {errors.city && (
            <FieldErrorMessage>{errors.city.message}</FieldErrorMessage>
          )}
        </Field>
        */}

        <Button type="submit" color="dark/white" className="w-full" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Complete Profile'}
        </Button>
      </FieldGroup>
    </form>
  );
}

