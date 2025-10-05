

export interface GenerationRequest {
  prompt: string;
  images: string[];
  provider: AIProvider;
  model: string;
  size?: string;
}

export interface GenerationResult {
  images: string[];
  text?: string;
}

export interface BackendResponse {
    images?: string[];
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
    modalities?: string[]; // 支持的模态枚举，例: ["text"], ["image"], ["text","image"]
    max_images: number;// 支持的最大输入图片数
    supported_sizes?: string[];// 支持的图像尺寸，留空表示不限预设
}


