import React, { useEffect, useState } from 'react';
import {
  Container,
  Typography,
  Box,
  ThemeProvider,
  createTheme,
  CssBaseline,
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

// 导入类型和服务
import type { GenerationRequest, GenerationResult, AIProvider, AIModel } from './types';
import { generateImage, fetchProviders } from './ai';
import ImageUpload from './components/ImageUpload';
import ImageViewer from './components/ImageViewer';

// 下载工具函数
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

// ImageGenerator 组件
interface ImageGeneratorProps {
  prompt: string;
  images: string[];
  onPromptChange: (prompt: string) => void;
  onGenerate: (result: GenerationResult) => void;
  onImagesChange: (images: string[]) => void;
  lastGeneratedImages?: string[];
  lastGenerationText?: string;
}

const ImageGenerator: React.FC<ImageGeneratorProps> = ({
  prompt,
  images,
  onPromptChange,
  onGenerate,
  onImagesChange,
  lastGeneratedImages = [],
  lastGenerationText,
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
        if (providerList.length > 0) {
          const provider = providerList[0];
          setSelectedProvider(provider);
          setSelectedModel(provider.models[0] ?? null);
        } else {
          setSelectedProvider(null);
          setSelectedModel(null);
          setSelectedSize('');
        }
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
        size: selectedModel?.inputs?.supported_sizes?.length ? selectedSize : undefined,
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
    if (!image) {
      return;
    }
    const timestamp = Date.now();
    downloadImage(image, `generated-${timestamp}-${index + 1}.png`);
  };

  const handleDownloadAll = () => {
    if (!lastGeneratedImages.length) {
      return;
    }
    const timestamp = Date.now();
    lastGeneratedImages.forEach((image, idx) => {
      downloadImage(image, `generated-${timestamp}-${idx + 1}.png`);
    });
  };

  // 打开图片查看器
  const handleImageClick = (index: number) => {
    setResultViewerIndex(index);
    setResultViewerOpen(true);
  };

  // 关闭图片查看器
  const handleCloseViewer = () => {
    setResultViewerOpen(false);
  };

  return (
    <Paper elevation={2} sx={{ p: 2 }}>
      {/* 标题部分 */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="h5" component="h2" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <AutoAwesome color="primary" />
          AI 图片生成器
        </Typography>
      </Box>

      {/* 图片上传区域 */}
      <Box sx={{ mb: 2 }}>
        <ImageUpload
          images={images}
          onImagesChange={onImagesChange}
          maxImages={5}
        />
      </Box>

      {/* 服务商和模型选择 */}
      <Box sx={{ mb: 2 }}>
        <Box sx={{ display: 'flex', gap: 2, flexDirection: { xs: 'column', sm: 'row' } }}>
          <FormControl fullWidth size="small">
            <InputLabel>AI 服务商</InputLabel>
            <Select
              value={selectedProvider?.id ?? ''}
              label="AI 服务商"
              disabled={isGenerating || providersLoading || providers.length === 0}
              onChange={(e) => {
                const provider = providers.find(p => p.id === e.target.value);
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
                const model = selectedProvider.models.find(m => m.id === e.target.value);
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
                    <Typography variant="body2">{model.name}</Typography>
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
          <Alert severity="error" sx={{ mt: 1 }}>
            {providerError}
          </Alert>
        )}
      </Box>

      {/* 提示词输入 */}
      <Box sx={{ mb: 2 }}>
        <TextField
          fullWidth
          multiline
          rows={3}
          label="创意描述"
          placeholder="详细描述你想要的图片..."
          value={prompt}
          onChange={(e) => onPromptChange(e.target.value)}
          disabled={isGenerating}
          helperText={
            selectedProvider && selectedModel
              ? `${prompt.length}/500 字符 | 使用 ${selectedProvider.name} - ${selectedModel.name}`
              : `${prompt.length}/500 字符`
          }
          size="small"
        />
      </Box>

      {/* 错误信息 */}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          <Box component="pre" sx={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit', m: 0 }}>
            {error}
          </Box>
        </Alert>
      )}

      {/* 生成按钮 */}
      <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
        <Button
          variant="contained"
          size="medium"
          fullWidth
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
            py: 1.5,
            background: 'linear-gradient(45deg, #FF6B6B 30%, #4ECDC4 90%)',
            '&:hover': {
              background: 'linear-gradient(45deg, #FF5252 30%, #26A69A 90%)',
            },
          }}
        >
          {isGenerating ? '创作中...' : '开始生成'}
        </Button>
      </Box>

      {/* 生成结果 */}
      {lastGeneratedImages.length > 0 && (
        <Card sx={{ mb: 2 }}>
          <Box sx={{ p: 1.5, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <ImageIcon color="success" fontSize="small" />
              <Typography variant="body2" fontWeight="medium">生成成功</Typography>
              <Chip label={`共 ${lastGeneratedImages.length} 张`} color="success" size="small" />
            </Box>
            <Button
              variant="contained"
              color="success"
              size="small"
              startIcon={<Download />}
              onClick={lastGeneratedImages.length > 1 ? handleDownloadAll : () => handleDownload(0)}
              sx={{ borderRadius: 2 }}
            >
              {lastGeneratedImages.length > 1 ? '下载全部' : '下载'}
            </Button>
          </Box>
          {lastGenerationText && (
            <Box sx={{ px: 2, pb: 1 }}>
              <Alert severity="info" sx={{ m: 0 }}>
                {lastGenerationText}
              </Alert>
            </Box>
          )}
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, px: 2, pb: 2 }}>
            {lastGeneratedImages.map((image, index) => (
              <Box
                key={image + index}
                sx={{
                  flex: '1 1 240px',
                  maxWidth: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 1,
                }}
              >
                <CardMedia
                  component="img"
                  image={image}
                  alt={`AI生成的图片 ${index + 1}`}
                  onClick={() => handleImageClick(index)}
                  sx={{
                    maxHeight: 420,
                    objectFit: 'contain',
                    bgcolor: 'grey.100',
                    cursor: 'pointer',
                    borderRadius: 1,
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

      {/* 图片查看器 */}
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

// 创建自定义主题
const theme = createTheme({
  palette: {
    primary: {
      main: '#3f51b5',
      light: '#7986cb',
      dark: '#303f9f',
    },
    secondary: {
      main: '#f50057',
      light: '#ff5983',
      dark: '#c51162',
    },
    background: {
      default: '#f5f5f5',
      paper: '#ffffff',
    },
  },
  typography: {
    h4: {
      fontWeight: 600,
    },
    h5: {
      fontWeight: 600,
    },
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 8,
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          textTransform: 'none',
          fontWeight: 600,
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: 6,
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 8,
        },
      },
    },
  },
});

function App() {
  const [prompt, setPrompt] = useState('');
  const [images, setImages] = useState<string[]>([]);
  const [lastGeneratedImages, setLastGeneratedImages] = useState<string[]>([]);
  const [lastGenerationText, setLastGenerationText] = useState<string>('');

  // 处理图片生成完成
  const handleGenerate = (result: GenerationResult) => {
    setLastGeneratedImages(result.images);
    setLastGenerationText(result.text ?? '');
  };



  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      
      {/* 主标题区域 */}
      <Box
        sx={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          color: 'white',
          py: 3,
          textAlign: 'center',
        }}
      >
        <Container maxWidth="md">
          <Typography variant="h4" component="h2" gutterBottom sx={{ fontWeight: 700 }}>
            AI图片生成器
          </Typography>
          <Typography variant="body1" sx={{ opacity: 0.9 }}>
            用文字描述你的想法，让AI为你创造精美的图片
          </Typography>
        </Container>
      </Box>

      {/* 主内容区域 */}
      <Container maxWidth="lg" sx={{ py: 2 }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <ImageGenerator
            prompt={prompt}
            images={images}
            onPromptChange={setPrompt}
            onGenerate={handleGenerate}
            onImagesChange={setImages}
            lastGeneratedImages={lastGeneratedImages}
            lastGenerationText={lastGenerationText || undefined}
          />
        </Box>
      </Container>

      {/* 底部信息 */}
      <Box
        component="footer"
        sx={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          color: 'white',
          py: 3,
          mt: 4,
        }}
      >
        <Container maxWidth="lg">
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(4, 1fr)' }, gap: 2 }}>
            <Box>
              <Typography variant="h6" gutterBottom>
                使用指南
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                <Typography variant="body2">1. 输入创意描述文本</Typography>
                <Typography variant="body2">2. 上传参考图片（可选）</Typography>
                <Typography variant="body2">3. 点击生成按钮创作</Typography>
                <Typography variant="body2">4. 下载</Typography>
              </Box>
            </Box>
            
            <Box>
              <Typography variant="h6" gutterBottom>
                ✨ 功能特色
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                <Typography variant="body2">🤖 AI智能生成</Typography>
                <Typography variant="body2">🖱️ 拖拽上传支持</Typography>
                <Typography variant="body2">🎨 即时图片预览</Typography>
                <Typography variant="body2">💾 一键下载保存</Typography>
              </Box>
            </Box>

            <Box>
              <Typography variant="h6" gutterBottom>
                🛠️ 技术栈
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                <Typography variant="body2">React + TypeScript</Typography>
                <Typography variant="body2">Material-UI</Typography>
                <Typography variant="body2">Gemini AI API</Typography>
                <Typography variant="body2">Vite 构建工具</Typography>
              </Box>
            </Box>

            <Box>
              <Typography variant="h6" gutterBottom>
                ⚙️ 环境配置
              </Typography>
              <Typography variant="body2" gutterBottom>
                需要配置 API 密钥：
              </Typography>
              <Box
                sx={{
                  bgcolor: 'rgba(0,0,0,0.3)',
                  p: 1,
                  borderRadius: 1,
                  fontFamily: 'monospace',
                  fontSize: '0.8rem',
                  mb: 1,
                }}
              >
                VITE_GEMINI_API_KEY<br/>
                VITE_OPENROUTER_API_KEY
              </Box>
              <Typography variant="caption" sx={{ display: 'block' }}>
                Gemini: Google AI Studio
              </Typography>
              <Typography variant="caption" sx={{ display: 'block' }}>
                OpenRouter: openrouter.ai
              </Typography>
            </Box>
          </Box>

          <Box sx={{ textAlign: 'center', mt: 2, pt: 2, borderTop: '1px solid rgba(255,255,255,0.2)' }}>
            <Typography variant="body2" sx={{ opacity: 0.8 }}>
              © 2025 AI图片生成器 - 让创意无限延伸 🚀
            </Typography>
          </Box>
        </Container>
      </Box>
    </ThemeProvider>
  );
}

export default App;
