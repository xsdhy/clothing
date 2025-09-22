import type { AIProvider, GenerationRequest } from '../types';

const LLM_BASE_PATH = '/api/llm';
const LLM_GENERATE_ENDPOINT = LLM_BASE_PATH;
const LLM_PROVIDERS_ENDPOINT = `${LLM_BASE_PATH}/providers`;

interface BackendResponse {
  image?: string;
  text?: string;
  error?: string;
}

interface ProvidersResponse {
  providers?: AIProvider[];
  error?: string;
}

const sanitizeImages = (images: string[] = []): string[] =>
  images
    .map((image) => (typeof image === 'string' ? image.trim() : ''))
    .filter((image): image is string => image.length > 0);

export const generateImage = async (request: GenerationRequest): Promise<string> => {
  const prompt = request.prompt.trim();
  if (!prompt) {
    throw new Error('请输入图片描述');
  }

  if (!request.provider?.id) {
    throw new Error('请选择服务商');
  }

  const payload = {
    prompt,
    images: sanitizeImages(request.images),
    provider: request.provider.id,
    model: request.model,
  };

  let response: Response;
  try {
    response = await fetch(LLM_GENERATE_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    throw new Error(`请求后端服务失败: ${error instanceof Error ? error.message : '未知错误'}`);
  }

  let body: BackendResponse | undefined;
  try {
    body = await response.json();
  } catch {
    body = undefined;
  }

  if (!response.ok) {
    const message = body?.error ?? `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (body?.error) {
    throw new Error(body.error);
  }

  if (body?.image) {
    return body.image;
  }

  if (body?.text) {
    return `data:text/plain;charset=utf-8,${encodeURIComponent(body.text)}`;
  }

  throw new Error('后端未返回有效的图片数据');
};

export const fetchProviders = async (): Promise<AIProvider[]> => {
  let response: Response;
  try {
    response = await fetch(LLM_PROVIDERS_ENDPOINT);
  } catch (error) {
    throw new Error(`请求模型列表失败: ${error instanceof Error ? error.message : '未知错误'}`);
  }

  let body: ProvidersResponse | undefined;
  try {
    body = await response.json();
  } catch {
    body = undefined;
  }

  if (!response.ok) {
    const message = body?.error ?? `服务请求失败: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  if (!body?.providers || !Array.isArray(body.providers)) {
    throw new Error('后端未返回有效的模型列表');
  }

  return body.providers;
};
