import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuth } from '../hooks/useAuth';
import { useCountries } from '../hooks/useCountries';
import { Button } from '../../../shared/components/ui/button';
import { Input, InputGroup } from '../../../shared/components/ui/input';
import { Select } from '../../../shared/components/ui/select';
import { Field, FieldGroup, Label, ErrorMessage as FieldErrorMessage } from '../../../shared/components/ui/fieldset';
import { Text } from '../../../shared/components/ui/text';
import { streamerRegisterSchema } from '../schemas';

export function StreamerRegisterForm() {
  const { register: registerStreamer } = useAuth();
  const navigate = useNavigate();
  const { countries, isLoading: countriesLoading } = useCountries();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm({
    resolver: zodResolver(streamerRegisterSchema),
    defaultValues: {
      profileType: 'streamer' as const,
    },
  });

  const selectedCountryId = watch('country_id');

  const onSubmit = async (data: unknown) => {
    setError(null);
    try {
      const validatedData = streamerRegisterSchema.parse(data);
      setIsLoading(true);
      const { confirmPassword, country_id, ...registerData } = validatedData;
      const result = await registerStreamer({ 
        ...registerData, 
        profileType: 'streamer',
        country_id: country_id && country_id !== '' ? country_id : undefined,
      });
      setIsLoading(false);
      if (result.success) {
        navigate('/');
      } else {
        setError(result.error || 'Registration failed. Please try again.');
      }
    } catch (err) {
      setIsLoading(false);
      setError('Validation failed. Please check your input.');
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
          <Label htmlFor="email">Email</Label>
          <InputGroup>
            <Input
              id="email"
              type="email"
              {...register('email')}
              data-invalid={errors.email ? '' : undefined}
            />
          </InputGroup>
          {errors.email && (
            <FieldErrorMessage>{errors.email.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="password">Password</Label>
          <InputGroup>
            <Input
              id="password"
              type="password"
              {...register('password')}
              data-invalid={errors.password ? '' : undefined}
            />
          </InputGroup>
          {errors.password && (
            <FieldErrorMessage>{errors.password.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="confirmPassword">Confirm Password</Label>
          <InputGroup>
            <Input
              id="confirmPassword"
              type="password"
              {...register('confirmPassword')}
              data-invalid={errors.confirmPassword ? '' : undefined}
            />
          </InputGroup>
          {errors.confirmPassword && (
            <FieldErrorMessage>{errors.confirmPassword.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="nickname">Nickname</Label>
          <InputGroup>
            <Input
              id="nickname"
              type="text"
              {...register('nickname')}
              data-invalid={errors.nickname ? '' : undefined}
            />
          </InputGroup>
          {errors.nickname && (
            <FieldErrorMessage>{errors.nickname.message}</FieldErrorMessage>
          )}
        </Field>

        <Field>
          <Label htmlFor="first_name">First Name</Label>
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
          <Label htmlFor="last_name">Last Name</Label>
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
        </Field>

        <Field>
          <Label htmlFor="country_id">Country (optional)</Label>
          <Select
            id="country_id"
            {...register('country_id')}
            data-invalid={errors.country_id ? '' : undefined}
            disabled={countriesLoading}
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

        {selectedCountryId && (
          <Field>
            <Label htmlFor="city">City (optional)</Label>
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
        )}

        <div className="flex flex-col items-stretch gap-3">
          <Button type="submit" disabled={isLoading} color="dark/zinc">
            {isLoading ? 'Registering...' : 'Create streamer account'}
          </Button>
        </div>
      </FieldGroup>
    </form>
  );
}
