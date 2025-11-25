import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { useAuth } from '../hooks/useAuth';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import type { RegisterData } from '../types';

export function RegisterForm() {
  const { register: registerUser } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<RegisterData & { confirmPassword: string }>();

  const password = watch('password');

  const onSubmit = async (data: RegisterData & { confirmPassword: string }) => {
    setError(null);
    if (data.password !== data.confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    setIsLoading(true);
    const { confirmPassword, ...registerData } = data;
    const success = await registerUser(registerData);
    setIsLoading(false);
    if (success) {
      navigate('/');
    } else {
      setError('Registration failed. Please try again.');
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {error && <ErrorMessage message={error} />}
      
      <div>
        <label htmlFor="name" className="block text-sm font-medium mb-1">
          Name (optional)
        </label>
        <input
          id="name"
          type="text"
          {...register('name')}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div>
        <label htmlFor="email" className="block text-sm font-medium mb-1">
          Email
        </label>
        <input
          id="email"
          type="email"
          {...register('email', { required: 'Email is required' })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        {errors.email && (
          <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium mb-1">
          Password
        </label>
        <input
          id="password"
          type="password"
          {...register('password', { 
            required: 'Password is required',
            minLength: { value: 6, message: 'Password must be at least 6 characters' }
          })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        {errors.password && (
          <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="confirmPassword" className="block text-sm font-medium mb-1">
          Confirm Password
        </label>
        <input
          id="confirmPassword"
          type="password"
          {...register('confirmPassword', { 
            required: 'Please confirm your password',
            validate: (value) => value === password || 'Passwords do not match'
          })}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        {errors.confirmPassword && (
          <p className="mt-1 text-sm text-red-600">{errors.confirmPassword.message}</p>
        )}
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:opacity-50"
      >
        {isLoading ? 'Registering...' : 'Register'}
      </button>
    </form>
  );
}

