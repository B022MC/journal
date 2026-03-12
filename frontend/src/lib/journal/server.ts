import { apiFetchWithSession } from "@/lib/api/server";
import type {
  ListPapersResponse,
  PaperItem,
  PaperRatingsResponse,
  SearchPapersResponse,
  UserInfo,
} from "@/lib/journal/contracts";

export interface PaperQuery {
  query?: string;
  zone?: string;
  discipline?: string;
  sort?: string;
  engine?: string;
  shadowCompare?: boolean;
  suggestionLimit?: number;
  page?: number;
  pageSize?: number;
}

export async function listPapers(query: PaperQuery = {}) {
  return apiFetchWithSession<ListPapersResponse>("/papers", {
    access: "optional",
    query: {
      discipline: query.discipline,
      page: query.page ?? 1,
      page_size: query.pageSize ?? 12,
      sort: normalizeListSort(query.sort),
      zone: query.zone,
    },
  });
}

export async function searchPapers(query: PaperQuery) {
  return apiFetchWithSession<SearchPapersResponse>("/papers/search", {
    access: "optional",
    query: {
      query: query.query,
      discipline: query.discipline,
      sort: normalizeSearchSort(query.sort),
      engine: query.engine,
      shadow_compare: query.shadowCompare || undefined,
      suggestion_limit: query.suggestionLimit ?? 6,
      page: query.page ?? 1,
      page_size: query.pageSize ?? 12,
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

function normalizeListSort(sort?: string) {
  switch (sort) {
    case "quality":
      return "highest_rated";
    case "newest":
      return "newest";
    default:
      return "newest";
  }
}

function normalizeSearchSort(sort?: string) {
  switch (sort) {
    case "newest":
    case "quality":
      return sort;
    default:
      return "relevance";
  }
}
