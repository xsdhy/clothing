

export interface GenerationRequest {
  prompt: string;
  images: string[];
  provider: AIProvider;
  model: string;
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
}




