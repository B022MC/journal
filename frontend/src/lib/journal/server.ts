import { apiFetchWithSession } from "@/lib/api/server";
import type {
  ListPapersResponse,
  PaperItem,
  PaperRatingsResponse,
  UserInfo,
} from "@/lib/journal/contracts";

export interface PaperQuery {
  query?: string;
  zone?: string;
  discipline?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
}

export async function listPapers(query: PaperQuery = {}) {
  const path = query.query ? "/papers/search" : "/papers";

  return apiFetchWithSession<ListPapersResponse>(path, {
    query: {
      discipline: query.discipline,
      page: query.page ?? 1,
      page_size: query.pageSize ?? 12,
      query: query.query,
      sort: query.sort,
      zone: query.zone,
    },
  });
}

export async function getPaper(id: string) {
  return apiFetchWithSession<PaperItem>(`/papers/${id}`);
}

export async function getPaperRatings(id: string, page = 1, pageSize = 6) {
  return apiFetchWithSession<PaperRatingsResponse>(`/papers/${id}/ratings`, {
    access: "optional",
    query: {
      page,
      page_size: pageSize,
    },
  });
}

export async function getCurrentUser() {
  return apiFetchWithSession<UserInfo>("/user/info", {
    access: "optional",
  });
}
