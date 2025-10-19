import React, { useEffect, useRef, useState } from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  TextField,
  Button,
  Alert,
  CircularProgress,
  Card,
  CardMedia,
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import {
  Send,
  Download,
  AutoAwesome,
  Image as ImageIcon,
} from '@mui/icons-material';
import { useLocation } from 'react-router-dom';

import type { GenerationRequest, GenerationResult, AIProvider, AIModel } from '../types';
import { generateImage, fetchProviders } from '../ai';
import ImageUpload from '../components/ImageUpload';
import ImageViewer from '../components/ImageViewer';

function downloadImage(dataUrl: string, filename: string = 'generated-image.png'): void {
  try {
    const link = document.createElement('a');
    link.href = dataUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  } catch (error) {
    console.error('Error downloading image:', error);
    alert('下载失败，请重试');
  }
}

interface ImageGeneratorProps {
  prompt: string;
  images: string[];
  onPromptChange: (prompt: string) => void;
  onGenerate: (result: GenerationResult) => void;
  onImagesChange: (images: string[]) => void;
  lastGeneratedImages?: string[];
  lastGenerationText?: string;
  initialProviderId?: string;
  initialModelId?: string;
  initialSize?: string;
  selectionKey?: string;
}

interface HistoryPrefillState {
  prompt?: string;
  inputImages?: string[];
  providerId?: string;
  modelId?: string;
  size?: string;
}

const ImageGenerator: React.FC<ImageGeneratorProps> = ({
  prompt,
  images,
  onPromptChange,
  onGenerate,
  onImagesChange,
  lastGeneratedImages = [],
  lastGenerationText,
  initialProviderId,
  initialModelId,
  initialSize,
  selectionKey,
}) => {
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [providersLoading, setProvidersLoading] = useState(false);
  const [providerError, setProviderError] = useState<string | null>(null);
  const [selectedProvider, setSelectedProvider] = useState<AIProvider | null>(null);
  const [selectedModel, setSelectedModel] = useState<AIModel | null>(null);
  const [selectedSize, setSelectedSize] = useState<string>('');
  const [resultViewerOpen, setResultViewerOpen] = useState(false);
  const [resultViewerIndex, setResultViewerIndex] = useState(0);
  const [selectionInitialized, setSelectionInitialized] = useState(false);

  useEffect(() => {
    setSelectionInitialized(false);
  }, [selectionKey]);

  useEffect(() => {
    const sizes = selectedModel?.inputs?.supported_sizes ?? [];
    if (sizes.length === 0) {
      setSelectedSize('');
      return;
    }
    setSelectedSize((current) => (sizes.includes(current) ? current : sizes[0]));
  }, [selectedModel]);

  useEffect(() => {
    let cancelled = false;

    const loadProviders = async () => {
      setProvidersLoading(true);
      setProviderError(null);
      try {
        const providerList = await fetchProviders();
        if (cancelled) {
          return;
        }
        setProviders(providerList);
        if (providerList.length === 0) {
          setSelectedProvider(null);
          setSelectedModel(null);
          setSelectedSize('');
        }
        setSelectionInitialized(false);
      } catch (err) {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : '获取模型列表失败';
          setProviderError(message);
        }
      } finally {
        if (!cancelled) {
          setProvidersLoading(false);
        }
      }
    };

    loadProviders();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (selectionInitialized || providers.length === 0) {
      return;
    }

    const provider =
      (initialProviderId ? providers.find((item) => item.id === initialProviderId) : undefined) ??
      providers[0] ??
      null;

    setSelectedProvider(provider);

    let model: AIModel | null = provider?.models?.[0] ?? null;
    if (provider && initialModelId) {
      const maybeModel = provider.models.find((item) => item.id === initialModelId);
      if (maybeModel) {
        model = maybeModel;
      }
    }

    setSelectedModel(model);

    let size = '';
    if (model?.inputs?.supported_sizes && model.inputs.supported_sizes.length > 0) {
      size = model.inputs.supported_sizes[0];
      if (initialSize && model.inputs.supported_sizes.includes(initialSize)) {
        size = initialSize;
      }
    }

    setSelectedSize(size);
    setSelectionInitialized(true);
  }, [providers, initialProviderId, initialModelId, initialSize, selectionInitialized]);

  const handleGenerate = async () => {
    if (!prompt.trim()) {
      setError('请输入提示词');
      return;
    }

    if (!selectedProvider || !selectedModel) {
      setError('请选择可用的模型');
      return;
    }

    setIsGenerating(true);
    setError(null);

    try {
      const request: GenerationRequest = {
        prompt: prompt.trim(),
        images,
        provider: selectedProvider,
        model: selectedModel.id,
        size: selectedModel.inputs?.supported_sizes?.length ? selectedSize : undefined,
      };

      const result = await generateImage(request);
      onGenerate(result);
      if (result.images.length > 0) {
        setResultViewerIndex(0);
      }
      setResultViewerOpen(false);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '生成失败，请重试';
      setError(errorMessage);
      console.error(`生成失败: ${errorMessage}`);
      console.error('Generation error:', err);
    } finally {
      setIsGenerating(false);
    }
  };

  const handleDownload = (index: number) => {
    const image = lastGeneratedImages[index];
    if (image) {
      downloadImage(image, `generated-image-${index + 1}.png`);
    }
  };

  const handleDownloadAll = () => {
    lastGeneratedImages.forEach((image, index) => {
      downloadImage(image, `generated-image-${index + 1}.png`);
    });
  };

  const handleOpenResultViewer = (index: number) => {
    setResultViewerIndex(index);
    setResultViewerOpen(true);
  };

  const handleCloseViewer = () => {
    setResultViewerOpen(false);
  };

  return (
    <Paper
      elevation={0}
      sx={{
        p: { xs: 2, sm: 3 },
        borderRadius: { xs: 2, sm: 3 },
        border: '1px solid',
        borderColor: 'divider',
        boxShadow: { xs: '0 12px 24px rgba(15,23,42,0.04)', sm: '0 18px 32px rgba(15,23,42,0.06)' },
        bgcolor: 'background.paper',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          mb: { xs: 2, sm: 3 },
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <AutoAwesome color="primary" />
          <Typography variant="h6" component="h2" sx={{ fontWeight: 600 }}>
            AI 图片生成器
          </Typography>
        </Box>
        {providersLoading && <CircularProgress size={20} />}
      </Box>

      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', md: 'minmax(112px, 168px) minmax(0,1fr)' },
          gap: { xs: 2, md: 3 },
          alignItems: 'stretch',
          mb: { xs: 2, sm: 3 },
        }}
      >
        <ImageUpload images={images} onImagesChange={onImagesChange} maxImages={5} variant="inline" />

        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', gap: { xs: 1.5, sm: 2 } }}>
          <TextField
            fullWidth
            multiline
            minRows={4}
            maxRows={12}
            label="创意描述"
            placeholder="详细描述你想要的图片..."
            value={prompt}
            onChange={(e) => onPromptChange(e.target.value)}
            disabled={isGenerating}
            InputProps={{
              sx: {
                alignItems: 'flex-start',
                px: { xs: 1.25, sm: 1.5 },
                py: { xs: 1, sm: 1.25 },
                '& .MuiInputBase-inputMultiline': {
                  fontSize: { xs: '0.9rem', sm: '0.95rem' },
                  lineHeight: 1.5,
                },
              },
            }}
            FormHelperTextProps={{
              sx: { display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: 1, mt: 0.5 },
            }}
            helperText={
              selectedProvider && selectedModel
                ? `${prompt.length}/500 字符 · ${selectedProvider.name} · ${selectedModel.name}`
                : `${prompt.length}/500 字符`
            }
          />
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {
                xs: 'repeat(auto-fit, minmax(140px, 1fr))',
                sm: 'repeat(auto-fit, minmax(180px, 1fr))',
              },
              gap: { xs: 1.5, sm: 1.5 },
            }}
          >
            <FormControl fullWidth size="small">
              <InputLabel>AI 服务商</InputLabel>
              <Select
                value={selectedProvider?.id ?? ''}
                label="AI 服务商"
                disabled={isGenerating || providersLoading || providers.length === 0}
                onChange={(e) => {
                  const provider = providers.find((p) => p.id === e.target.value);
                  if (!provider) {
                    return;
                  }
                  setSelectedProvider(provider);
                  setSelectedModel(provider.models[0] ?? null);
                  setSelectedSize(provider.models[0]?.inputs?.supported_sizes?.[0] ?? '');
                }}
              >
                {providersLoading && providers.length === 0 && (
                  <MenuItem value="" disabled>
                    加载中...
                  </MenuItem>
                )}
                {providers.map((provider) => (
                  <MenuItem key={provider.id} value={provider.id}>
                    {provider.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl fullWidth size="small">
              <InputLabel>模型</InputLabel>
              <Select
                value={selectedModel?.id ?? ''}
                label="模型"
                disabled={
                  isGenerating ||
                  providersLoading ||
                  !selectedProvider ||
                  (selectedProvider?.models?.length ?? 0) === 0
                }
                onChange={(e) => {
                  if (!selectedProvider) {
                    return;
                  }
                  const model = selectedProvider.models.find((m) => m.id === e.target.value);
                  if (!model) {
                    return;
                  }
                  setSelectedModel(model);
                  setSelectedSize(model.inputs?.supported_sizes?.[0] ?? '');
                }}
              >
                {(selectedProvider?.models ?? []).map((model) => (
                  <MenuItem key={model.id} value={model.id}>
                    <Box>
                      <Typography variant="body2" sx={{ fontWeight: 500 }}>
                        {model.name}
                      </Typography>
                      {model.description && (
                        <Typography variant="caption" color="text.secondary">
                          {model.description}
                        </Typography>
                      )}
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {selectedModel?.inputs?.supported_sizes && selectedModel.inputs.supported_sizes.length > 0 && (
              <FormControl fullWidth size="small">
                <InputLabel>图片尺寸</InputLabel>
                <Select
                  value={selectedSize}
                  label="图片尺寸"
                  disabled={isGenerating}
                  onChange={(e) => setSelectedSize(e.target.value)}
                >
                  {selectedModel.inputs.supported_sizes.map((sizeOption) => (
                    <MenuItem key={sizeOption} value={sizeOption}>
                      {sizeOption}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}
          </Box>

          {providerError && (
            <Alert severity="error" sx={{ m: 0 }}>
              {providerError}
            </Alert>
          )}
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: { xs: 2, sm: 3 } }}>
          <Box component="pre" sx={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit', m: 0 }}>
            {error}
          </Box>
        </Alert>
      )}

      <Box
        sx={{
          display: 'flex',
          flexDirection: { xs: 'column', sm: 'row' },
          justifyContent: { xs: 'stretch', sm: 'flex-end' },
          alignItems: { xs: 'stretch', sm: 'center' },
          gap: 1,
          mb: { xs: 2, sm: 3 },
        }}
      >
        <Button
          variant="contained"
          size="medium"
          onClick={handleGenerate}
          disabled={
            isGenerating ||
            !prompt.trim() ||
            providersLoading ||
            !selectedProvider ||
            !selectedModel
          }
          startIcon={isGenerating ? <CircularProgress size={18} color="inherit" /> : <Send />}
          sx={{
            width: { xs: '100%', sm: 'auto' },
            minWidth: { sm: 200 },
            py: 1.4,
            fontWeight: 600,
            borderRadius: 999,
            background: 'linear-gradient(45deg, #FF6B6B 30%, #4ECDC4 90%)',
            '&:hover': {
              background: 'linear-gradient(45deg, #FF5252 30%, #26A69A 90%)',
            },
          }}
        >
          {isGenerating ? '创作中...' : '开始生成'}
        </Button>
      </Box>

      {lastGeneratedImages.length > 0 && (
        <Card
          sx={{
            mb: { xs: 2, sm: 3 },
            borderRadius: { xs: 2, sm: 3 },
            boxShadow: '0 10px 28px rgba(15,23,42,0.08)',
          }}
        >
          <Box
            sx={{
              p: { xs: 1.5, sm: 2 },
              display: 'flex',
              flexDirection: { xs: 'column', sm: 'row' },
              gap: { xs: 1, sm: 2 },
              alignItems: { xs: 'flex-start', sm: 'center' },
              justifyContent: 'space-between',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <ImageIcon color="success" fontSize="small" />
              <Typography variant="body2" fontWeight="medium">
                生成成功
              </Typography>
              <Chip label={`共 ${lastGeneratedImages.length} 张`} color="success" size="small" />
            </Box>
            <Button
              variant="contained"
              color="success"
              size="small"
              startIcon={<Download />}
              onClick={lastGeneratedImages.length > 1 ? handleDownloadAll : () => handleDownload(0)}
              sx={{ borderRadius: 999, px: 2.5 }}
            >
              {lastGeneratedImages.length > 1 ? '下载全部' : '下载'}
            </Button>
          </Box>
          {lastGenerationText && (
            <Box sx={{ px: { xs: 1.5, sm: 2 }, pb: { xs: 1.5, sm: 2 } }}>
              <Alert severity="info" sx={{ m: 0 }}>
                {lastGenerationText}
              </Alert>
            </Box>
          )}
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: 'repeat(auto-fit, minmax(160px, 1fr))', sm: 'repeat(auto-fit, minmax(220px, 1fr))' },
              gap: { xs: 1.5, sm: 2 },
              px: { xs: 1.5, sm: 2 },
              pb: { xs: 1.5, sm: 2 },
            }}
          >
            {lastGeneratedImages.map((image, index) => (
              <Box
                key={image + index}
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 1,
                }}
              >
                <CardMedia
                  component="img"
                  image={image}
                  alt={`AI生成的图片 ${index + 1}`}
                  onClick={() => handleOpenResultViewer(index)}
                  sx={{
                    width: '100%',
                    maxHeight: 420,
                    objectFit: 'cover',
                    bgcolor: 'grey.100',
                    cursor: 'pointer',
                    borderRadius: 2,
                    transition: 'all 0.2s',
                    '&:hover': {
                      filter: 'brightness(1.05)',
                      transform: 'scale(1.01)',
                    },
                  }}
                />
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<Download />}
                  onClick={() => handleDownload(index)}
                  sx={{ alignSelf: 'flex-end' }}
                >
                  下载第 {index + 1} 张
                </Button>
              </Box>
            ))}
          </Box>
        </Card>
      )}

      <ImageViewer
        open={resultViewerOpen}
        onClose={handleCloseViewer}
        imageUrl={lastGeneratedImages[resultViewerIndex] || ''}
        title={`AI生成的图片 ${resultViewerIndex + 1}/${lastGeneratedImages.length}`}
        showDownload
        onDownload={() => handleDownload(resultViewerIndex)}
      />

      {lastGeneratedImages.length === 0 && lastGenerationText && (
        <Alert severity="info" sx={{ mt: 2 }}>
          {lastGenerationText}
        </Alert>
      )}
    </Paper>
  );
};

const CustomImageGenerationPage: React.FC = () => {
  const location = useLocation();
  const [prompt, setPrompt] = useState('');
  const [images, setImages] = useState<string[]>([]);
  const [lastGeneratedImages, setLastGeneratedImages] = useState<string[]>([]);
  const [lastGenerationText, setLastGenerationText] = useState<string | null>(null);
  const [initialSelection, setInitialSelection] = useState<{
    providerId?: string;
    modelId?: string;
    size?: string;
  }>({});
  const [selectionKey, setSelectionKey] = useState<string>('');
  const lastAppliedPrefillKeyRef = useRef<string | null>(null);

  useEffect(() => {
    const state = location.state as HistoryPrefillState | null | undefined;

    if (!state) {
      return;
    }

    if (lastAppliedPrefillKeyRef.current === location.key) {
      return;
    }

    setPrompt(state.prompt ?? '');
    setImages(state.inputImages ?? []);
    setInitialSelection({
      providerId: state.providerId,
      modelId: state.modelId,
      size: state.size,
    });
    setSelectionKey(location.key);
    setLastGeneratedImages([]);
    setLastGenerationText(null);
    lastAppliedPrefillKeyRef.current = location.key;
  }, [location]);

  const handleGenerate = (result: GenerationResult) => {
    setLastGeneratedImages(result.images);
    setLastGenerationText(result.text ?? null);
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100%' }}>
      <Container maxWidth="md" sx={{ py: { xs: 2, sm: 3 }, flex: 1 }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: { xs: 2, sm: 3 } }}>
          <ImageGenerator
            prompt={prompt}
            images={images}
            onPromptChange={setPrompt}
            onGenerate={handleGenerate}
            onImagesChange={setImages}
            lastGeneratedImages={lastGeneratedImages}
            lastGenerationText={lastGenerationText ?? undefined}
            initialProviderId={initialSelection.providerId}
            initialModelId={initialSelection.modelId}
            initialSize={initialSelection.size}
            selectionKey={selectionKey}
          />
        </Box>
      </Container>

    </Box>
  );
};

export default CustomImageGenerationPage;
