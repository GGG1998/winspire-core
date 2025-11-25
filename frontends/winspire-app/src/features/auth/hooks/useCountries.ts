import { useEffect, useState } from 'react';
import { countriesApi, type Country } from '../api/countriesApi';

export function useCountries() {
  const [countries, setCountries] = useState<Country[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetchCountries() {
      try {
        setIsLoading(true);
        setError(null);
        const data = await countriesApi.getAll();
        if (!cancelled) {
          setCountries(data);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error('Failed to fetch countries'));
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    fetchCountries();

    return () => {
      cancelled = true;
    };
  }, []);

  return { countries, isLoading, error };
}

