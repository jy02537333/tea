import { useQuery } from '@tanstack/react-query';
import type { Store } from '../services/stores';
import { getStores } from '../services/stores';
import type { PaginatedResult } from '../services/api';

export function useStores(options?: { enabled?: boolean; page?: number; limit?: number }) {
  const page = options?.page ?? 1;
  const limit = options?.limit ?? 500;
  const enabled = options?.enabled ?? true;

  return useQuery<PaginatedResult<Store>>({
    queryKey: ['stores', page, limit],
    queryFn: () => getStores({ page, limit }),
    enabled,
    staleTime: 1000 * 60 * 10,
    gcTime: 1000 * 60 * 60,
    refetchOnWindowFocus: false,
    refetchOnMount: false,
    refetchOnReconnect: false,
  });
}
