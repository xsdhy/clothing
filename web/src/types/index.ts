export interface GenerationInputs {
  images: string[];
  videos?: string[];
}

export interface GenerationOptions {
  size?: string;
  duration?: number;
}

export interface GenerationRequest {
  prompt: string;
  inputs: GenerationInputs;
  options?: GenerationOptions;
  provider: AIProvider;
  model: string;
  tag_ids?: number[];
}

export interface GenerationResult {
  outputs: string[];
  text?: string;
}

export interface BackendResponse {
  outputs?: string[];
  text?: string;
  error?: string;
}

export interface ProvidersResponse {
  providers?: AIProvider[];
  error?: string;
}

export interface AIProvider {
  id: string;
  name: string;
  models: AIModel[];
}

export interface AIModel {
  id: string;
  name: string;
  description?: string;

  inputs?: AIModelInput;
}

export interface AIModelInput {
  input_modalities?: string[]; // 支持的输入模态
  output_modalities?: string[]; // 支持的输出模态
  max_images: number; // 支持的最大输入图片数
  supported_sizes?: string[]; // 支持的图像尺寸，留空表示不限预设
  supported_durations?: number[]; // 视频时长（秒）
  default_size?: string;
  default_duration?: number;
}

export interface UserSummary {
  id: number;
  email: string;
  display_name?: string;
  role: string;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: UserSummary;
}

export interface AuthStatusResponse {
  has_user: boolean;
}

export interface UsageImage {
  path: string;
  url: string;
}

export interface Tag {
  id: number;
  name: string;
  usage_count?: number;
  created_at?: string;
  updated_at?: string;
}

export interface UsageRecord {
  id: number;
  provider_id: string;
  model_id: string;
  prompt: string;
  size?: string;
  output_text?: string;
  error_message?: string;
  created_at: string;
  input_images: UsageImage[];
  output_images: UsageImage[];
  user?: UserSummary;
  tags: Tag[];
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total: number;
}

export interface UsageRecordListResponse {
  records: UsageRecord[];
  meta: PaginationMeta;
}

export interface UsageRecordDetailResponse {
  record: UsageRecord;
}

export interface UserListResponse {
  users: UserSummary[];
  meta: PaginationMeta;
}

export interface ProviderModelAdmin {
  model_id: string;
  name: string;
  description?: string;
  price?: string;
  max_images?: number;
  input_modalities?: string[];
  output_modalities?: string[];
  supported_sizes?: string[];
  supported_durations?: number[];
  default_size?: string;
  default_duration?: number;
  settings?: Record<string, unknown>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProviderAdmin {
  id: string;
  name: string;
  driver: string;
  description?: string;
  base_url?: string;
  config?: Record<string, unknown>;
  is_active: boolean;
  has_api_key: boolean;
  created_at: string;
  updated_at: string;
  models?: ProviderModelAdmin[];
}

export interface ProviderCreatePayload {
  id: string;
  name: string;
  driver: string;
  description?: string;
  api_key?: string;
  base_url?: string;
  config?: Record<string, unknown>;
  is_active?: boolean;
}

export interface ProviderUpdatePayload {
  name?: string;
  driver?: string;
  description?: string;
  api_key?: string;
  base_url?: string;
  config?: Record<string, unknown>;
  is_active?: boolean;
}

export interface ProviderModelCreatePayload {
  model_id: string;
  name: string;
  description?: string;
  price?: string;
  max_images?: number;
  input_modalities?: string[];
  output_modalities?: string[];
  supported_sizes?: string[];
  supported_durations?: number[];
  default_size?: string;
  default_duration?: number;
  settings?: Record<string, unknown>;
  is_active?: boolean;
}

export interface ProviderModelUpdatePayload {
  name?: string;
  description?: string;
  price?: string;
  max_images?: number;
  input_modalities?: string[];
  output_modalities?: string[];
  supported_sizes?: string[];
  supported_durations?: number[];
  default_size?: string;
  default_duration?: number;
  settings?: Record<string, unknown>;
  is_active?: boolean;
}

export interface TagListResponse {
  tags: Tag[];
}

export interface TagDetailResponse {
  tag: Tag;
}
