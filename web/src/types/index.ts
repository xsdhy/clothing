

export interface GenerationRequest {
  prompt: string;
  images: string[];
  provider: AIProvider;
  model: string;
  size?: string;
  timeoutMs?: number;
}

export interface GenerationResult {
  images: string[];
  text?: string;
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
  image_sizes?: string[];
}


