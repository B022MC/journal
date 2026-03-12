export interface PaperItem {
  id: number;
  title: string;
  title_en: string;
  abstract: string;
  abstract_en: string;
  content?: string;
  author_id: number;
  author_name: string;
  discipline: string;
  zone: string;
  shit_score: number;
  avg_rating: number;
  rating_count: number;
  view_count: number;
  controversy_index: number;
  doi: string;
  keywords: string;
  file_path: string;
  status: number;
  degradation_level: number;
  created_at: number;
  promoted_at: number;
}

export interface ListPapersResponse {
  items: PaperItem[];
  total: number;
}

export interface SearchMeta {
  engine: string;
  used_fallback: boolean;
  fallback_reason: string;
  shadow_compared: boolean;
  indexed_docs: number;
  indexed_terms: number;
  index_signature: string;
  expanded_terms: string[];
}

export interface SearchPapersResponse {
  items: PaperItem[];
  total: number;
  suggestions: string[];
  meta: SearchMeta;
}

export interface RatingItem {
  id: number;
  paper_id: number;
  user_id: number;
  username: string;
  nickname: string;
  score: number;
  comment: string;
  created_at: number;
}

export interface PaperRatingsResponse {
  items: RatingItem[];
  total: number;
  avg_score: number;
}

export interface UserInfo {
  id: number;
  username: string;
  email: string;
  nickname: string;
  avatar: string;
  role: number;
  contribution_score: string;
  created_at: number;
  admin_permissions: string[];
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expire_at: number;
  user_info: UserInfo;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  nickname?: string;
}

export interface IdResponse {
  id: number;
}

export interface CommonResponse {
  success: boolean;
  message: string;
}

export interface FlagStatusResponse {
  exists: boolean;
  target_type: string;
  target_id: number;
  flag_count: number;
  pending_count: number;
  weighted_sum: number;
  quorum: number;
  degradation_level: number;
}

export interface FlagActionResponse {
  success: boolean;
  message: string;
  flag_id: number;
  status: FlagStatusResponse;
}
