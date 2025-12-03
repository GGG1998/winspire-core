import { supabase } from '../../../shared/api/supabase';

export interface Country {
  id: string;
  name: string;
  iso_code: string;
  created_at: string;
}

export const countriesApi = {
  getAll: async (): Promise<Country[]> => {
    const { data, error } = await supabase
      .from('countries')
      .select('id, name, iso_code, created_at')
      .order('name', { ascending: true });

    if (error) {
      console.error('Error fetching countries:', error);
      throw error;
    }

    return data || [];
  },
};






