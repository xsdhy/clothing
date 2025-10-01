import React, { useState, useRef, useCallback, useEffect } from 'react';
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
  IconButton,
  Dialog,
  DialogContent,
  Zoom,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import {
  Send,
  Download,
  AutoAwesome,
  Image as ImageIcon,
  CloudUpload,
  Delete,
  DragIndicator,
  Close,
  ZoomIn,
  ZoomOut,
  ZoomOutMap,
} from '@mui/icons-material';

// 导入类型和服务
import type { GenerationRequest, GenerationResult, AIProvider, AIModel } from './types';
import { generateImage, fetchProviders } from './ai';

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

// ImageViewer 组件
interface ImageViewerProps {
  open: boolean;
  onClose: () => void;
  imageUrl: string;
  title?: string;
  showDownload?: boolean;
  onDownload?: () => void;
}

const ImageViewer: React.FC<ImageViewerProps> = ({
  open,
  onClose,
  imageUrl,
  title,
  showDownload = false,
  onDownload,
}) => {
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [lastPosition, setLastPosition] = useState({ x: 0, y: 0 });
  
  const theme = useTheme();
  const fullScreen = useMediaQuery(theme.breakpoints.down('md'));
  const dialogContentRef = useRef<HTMLDivElement>(null);

  // 重置状态
  const resetView = () => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
  };

  // 关闭对话框时重置状态
  const handleClose = () => {
    resetView();
    onClose();
  };

  // 缩放控制
  const handleZoomIn = () => {
    setScale(prev => Math.min(prev * 1.2, 5));
  };

  const handleZoomOut = () => {
    setScale(prev => Math.max(prev / 1.2, 0.1));
  };

  const handleResetZoom = () => {
    resetView();
  };

  // 鼠标拖拽
  const handleMouseDown = (e: React.MouseEvent) => {
    if (scale > 1) {
      setIsDragging(true);
      setLastPosition({ x: e.clientX - position.x, y: e.clientY - position.y });
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging && scale > 1) {
      setPosition({
        x: e.clientX - lastPosition.x,
        y: e.clientY - lastPosition.y,
      });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  // 使用 useEffect 来处理滚轮事件
  useEffect(() => {
    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      const delta = e.deltaY > 0 ? 0.9 : 1.1;
      setScale(prev => Math.max(0.1, Math.min(5, prev * delta)));
    };

    const dialogContent = dialogContentRef.current;
    if (dialogContent && open) {
      dialogContent.addEventListener('wheel', handleWheel, { passive: false });
      return () => {
        dialogContent.removeEventListener('wheel', handleWheel);
      };
    }
  }, [open]);

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth={false}
      fullWidth
      fullScreen={fullScreen}
      PaperProps={{
        sx: {
          bgcolor: 'rgba(0, 0, 0, 0.9)',
          backgroundImage: 'none',
          boxShadow: 'none',
          margin: 0,
          maxHeight: '100vh',
          maxWidth: '100vw',
        },
      }}
    >
      {/* 顶部工具栏 */}
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          zIndex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          p: 2,
          background: 'linear-gradient(180deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
        }}
      >
        {/* 左侧：标题 */}
        <Box sx={{ flex: 1 }}>
          {title && (
            <Typography variant="h6" sx={{ color: 'white', opacity: 0.9 }}>
              {title}
            </Typography>
          )}
        </Box>

        {/* 右侧：工具按钮 */}
        <Box sx={{ display: 'flex', gap: 1 }}>
          <IconButton
            onClick={handleZoomOut}
            sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }}
            size="small"
            >
            <ZoomOut />
          </IconButton>
          <IconButton
            onClick={handleResetZoom}
            sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }}
            size="small"
          >
            <ZoomOutMap />
          </IconButton>
          <IconButton
            onClick={handleZoomIn}
            sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }}
            size="small"
          >
            <ZoomIn />
          </IconButton>
          {showDownload && onDownload && (
            <IconButton
              onClick={onDownload}
              sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }}
              size="small"
            >
              <Download />
            </IconButton>
          )}
          <IconButton
            onClick={handleClose}
            sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }}
            size="small"
          >
            <Close />
          </IconButton>
        </Box>
      </Box>

      {/* 图片内容区域 */}
      <DialogContent
        ref={dialogContentRef}
        sx={{
          p: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          overflow: 'hidden',
          cursor: scale > 1 ? (isDragging ? 'grabbing' : 'grab') : 'default',
        }}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      >
        <Zoom in={open} timeout={300}>
          <Box
            component="img"
            src={imageUrl}
            alt={title || '图片预览'}
            sx={{
              maxWidth: scale === 1 ? '90vw' : 'none',
              maxHeight: scale === 1 ? '90vh' : 'none',
              width: scale === 1 ? 'auto' : undefined,
              height: scale === 1 ? 'auto' : undefined,
              transform: `scale(${scale}) translate(${position.x / scale}px, ${position.y / scale}px)`,
              transformOrigin: 'center',
              transition: isDragging ? 'none' : 'transform 0.1s ease-out',
              userSelect: 'none',
              pointerEvents: 'none',
            }}
          />
        </Zoom>
      </DialogContent>

      {/* 底部状态栏 */}
      <Box
        sx={{
          position: 'absolute',
          bottom: 0,
          left: 0,
          right: 0,
          zIndex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          p: 2,
          background: 'linear-gradient(0deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
        }}
      >
        <Typography variant="body2" sx={{ color: 'white', opacity: 0.7 }}>
          缩放: {Math.round(scale * 100)}% | 滚轮缩放 | 拖拽移动
        </Typography>
      </Box>
    </Dialog>
  );
};

// ImageUpload 组件
interface ImageUploadProps {
  images: string[];
  onImagesChange: (images: string[]) => void;
  maxImages?: number;
}

const ImageUpload: React.FC<ImageUploadProps> = ({
  images,
  onImagesChange,
  maxImages = 5,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const [viewerOpen, setViewerOpen] = useState(false);
  const [currentImageIndex, setCurrentImageIndex] = useState(0);
  
  // 拖拽排序相关状态
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  
  // 粘贴上传相关状态
  const [isPasteActive, setIsPasteActive] = useState(false);

  const processFiles = useCallback((files: File[]) => {
    if (files.length === 0) return;
    
    // 检查文件数量限制
    if (images.length + files.length > maxImages) {
      alert(`最多只能上传 ${maxImages} 张图片`);
      return;
    }

    // 转换为base64
    Promise.all(
      files.map((file) => {
        return new Promise<string>((resolve, reject) => {
          // 检查文件类型
          if (!file.type.startsWith('image/')) {
            reject(new Error('请选择图片文件'));
            return;
          }
          
          // 检查文件大小（限制为5MB）
          if (file.size > 5 * 1024 * 1024) {
            reject(new Error('图片文件不能超过 5MB'));
            return;
          }

          const reader = new FileReader();
          reader.onload = (e) => {
            resolve(e.target?.result as string);
          };
          reader.onerror = () => reject(new Error('文件读取失败'));
          reader.readAsDataURL(file);
        });
      })
    )
      .then((base64Images) => {
        onImagesChange([...images, ...base64Images]);
      })
      .catch((error) => {
        alert(error.message);
      });
  }, [images, onImagesChange, maxImages]);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || []);
    processFiles(files);

    // 清空input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // 拖拽事件处理
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);

    const files = Array.from(e.dataTransfer.files);
    processFiles(files);
  }, [processFiles]);

  const removeImage = (index: number) => {
    const newImages = [...images];
    newImages.splice(index, 1);
    onImagesChange(newImages);
  };

  // 打开图片查看器
  const handleImageClick = (index: number) => {
    setCurrentImageIndex(index);
    setViewerOpen(true);
  };

  // 关闭图片查看器
  const handleCloseViewer = () => {
    setViewerOpen(false);
  };

  // 拖拽排序事件处理
  const handleImageDragStart = (e: React.DragEvent, index: number) => {
    setDraggedIndex(index);
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/html', e.currentTarget.outerHTML);
    e.dataTransfer.setDragImage(e.currentTarget as HTMLElement, 40, 40);
  };

  const handleImageDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    
    if (draggedIndex !== null && draggedIndex !== index) {
      setDragOverIndex(index);
    }
  };

  const handleImageDragLeave = () => {
    setDragOverIndex(null);
  };

  const handleImageDrop = (e: React.DragEvent, dropIndex: number) => {
    e.preventDefault();
    
    if (draggedIndex !== null && draggedIndex !== dropIndex) {
      const newImages = [...images];
      const draggedImage = newImages[draggedIndex];
      
      // 移除被拖拽的元素
      newImages.splice(draggedIndex, 1);
      
      // 在新位置插入
      newImages.splice(dropIndex, 0, draggedImage);
      
      onImagesChange(newImages);
    }
    
    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  const handleImageDragEnd = () => {
    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  // 处理粘贴事件
  const handlePaste = useCallback((e: ClipboardEvent) => {
    // 检查是否在输入框中，如果是则不处理
    const target = e.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      return;
    }

    const clipboardData = e.clipboardData;
    if (!clipboardData) return;

    const items = Array.from(clipboardData.items);
    const imageItems = items.filter(item => item.type.startsWith('image/'));

    if (imageItems.length === 0) return;

    e.preventDefault();
    setIsPasteActive(true);

    // 处理图片文件
    const imageFiles: File[] = [];
    imageItems.forEach(item => {
      const file = item.getAsFile();
      if (file) {
        imageFiles.push(file);
      }
    });

    if (imageFiles.length > 0) {
      processFiles(imageFiles);
    }

    // 重置粘贴状态
    setTimeout(() => {
      setIsPasteActive(false);
    }, 500);
  }, [processFiles]);

  // 添加和移除粘贴事件监听器
  useEffect(() => {
    document.addEventListener('paste', handlePaste);
    
    return () => {
      document.removeEventListener('paste', handlePaste);
    };
  }, [handlePaste]);

  return (
    <Paper 
      elevation={2} 
      sx={{ 
        p: 2, 
        position: 'relative',
        border: isPasteActive ? '2px solid' : '2px solid transparent',
        borderColor: isPasteActive ? 'success.main' : 'transparent',
        transition: 'all 0.3s',
      }}
    >
      {/* 粘贴状态提示 */}
      {isPasteActive && (
        <Box
          sx={{
            position: 'absolute',
            top: 8,
            right: 8,
            bgcolor: 'success.main',
            color: 'white',
            px: 1,
            py: 0.5,
            borderRadius: 1,
            fontSize: '0.75rem',
            fontWeight: 'bold',
            zIndex: 10,
          }}
        >
          📋 粘贴上传中...
        </Box>
      )}

      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        multiple
        onChange={handleFileChange}
        style={{ display: 'none' }}
      />

      {/* 图片预览与上传区域一行显示 */}
      {images.length === 0 ? (
        /* 没有图片时显示完整上传区域 */
        <Box
          sx={{
            border: '2px dashed',
            borderColor: isDragOver ? 'primary.main' : 'grey.300',
            borderRadius: 2,
            p: 2.5,
            textAlign: 'center',
            cursor: 'pointer',
            bgcolor: isDragOver ? 'primary.50' : 'grey.50',
            transition: 'all 0.3s',
            position: 'relative',
            '&:hover': {
              borderColor: 'primary.light',
              bgcolor: 'primary.50',
            },
          }}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
        >
          {isDragOver && (
            <Box
              sx={{
                position: 'absolute',
                inset: 0,
                bgcolor: 'primary.100',
                border: '2px dashed',
                borderColor: 'primary.main',
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 1,
              }}
            >
              <Box sx={{ textAlign: 'center', color: 'primary.dark' }}>
                <DragIndicator sx={{ fontSize: 40, mb: 1 }} />
                <Typography variant="body1">松开鼠标即可上传</Typography>
              </Box>
            </Box>
          )}
          
          <Box>
            <CloudUpload sx={{ fontSize: 48, color: 'primary.main', mb: 1.5 }} />
            <Typography variant="body1" gutterBottom>
              拖拽图片到这里，或点击选择文件
            </Typography>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              支持 JPG、PNG 格式，单个文件不超过 5MB，最多上传 {maxImages} 张
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.75rem' }}>
              💡 提示：也可以使用 Ctrl+V (Cmd+V) 粘贴图片
            </Typography>
          </Box>
        </Box>
      ) : (
        /* 有图片时显示图片列表和紧贴的上传区域 */
        <Box sx={{ display: 'flex', gap: 1.5, alignItems: 'flex-start' }}>
          {/* 已上传图片预览 */}
          {images.map((image, index) => (
            <Box key={index} sx={{ flexShrink: 0 }}>
              <Card 
                draggable
                onDragStart={(e) => handleImageDragStart(e, index)}
                onDragOver={(e) => handleImageDragOver(e, index)}
                onDragLeave={handleImageDragLeave}
                onDrop={(e) => handleImageDrop(e, index)}
                onDragEnd={handleImageDragEnd}
                sx={{ 
                  position: 'relative',
                  width: '80px',
                  height: '80px',
                  cursor: draggedIndex === index ? 'grabbing' : 'grab',
                  opacity: draggedIndex === index ? 0.5 : 1,
                  transform: dragOverIndex === index ? 'scale(1.1)' : 'scale(1)',
                  border: dragOverIndex === index ? '2px solid' : '1px solid transparent',
                  borderColor: dragOverIndex === index ? 'primary.main' : 'transparent',
                  '&:hover .delete-btn': {
                    opacity: 1,
                  },
                  '&:hover .drag-indicator': {
                    opacity: 1,
                  },
                  '&:hover': {
                    transform: dragOverIndex === index ? 'scale(1.1)' : 'scale(1.05)',
                    boxShadow: 2,
                  },
                  transition: 'all 0.2s',
                }}
                onClick={() => draggedIndex === null && handleImageClick(index)}
              >
                <CardMedia
                  component="img"
                  width="80"
                  height="80"
                  image={image}
                  alt={`参考图片 ${index + 1}`}
                  sx={{ objectFit: 'cover' }}
                />
                {/* 拖拽指示器 */}
                <Box
                  className="drag-indicator"
                  sx={{
                    position: 'absolute',
                    top: 2,
                    left: 2,
                    bgcolor: 'rgba(0,0,0,0.7)',
                    color: 'white',
                    borderRadius: 0.5,
                    width: '20px',
                    height: '20px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    opacity: draggedIndex === index ? 1 : 0,
                    transition: 'opacity 0.2s',
                  }}
                >
                  <DragIndicator sx={{ fontSize: '12px' }} />
                </Box>
                
                <IconButton
                  className="delete-btn"
                  size="small"
                  color="error"
                  onClick={(e) => {
                    e.stopPropagation();
                    removeImage(index);
                  }}
                  sx={{
                    position: 'absolute',
                    top: 2,
                    right: 2,
                    bgcolor: 'error.main',
                    color: 'white',
                    opacity: 0,
                    transition: 'opacity 0.2s',
                    width: '20px',
                    height: '20px',
                    '&:hover': {
                      bgcolor: 'error.dark',
                    },
                  }}
                >
                  <Delete sx={{ fontSize: '14px' }} />
                </IconButton>
                <Box
                  sx={{
                    position: 'absolute',
                    bottom: 2,
                    left: 2,
                    bgcolor: 'rgba(0,0,0,0.7)',
                    color: 'white',
                    px: 0.5,
                    py: 0.25,
                    borderRadius: 0.5,
                    fontSize: '0.65rem',
                    lineHeight: 1,
                  }}
                >
                  #{index + 1}
                </Box>
              </Card>
            </Box>
          ))}

          {/* 拖拽上传区域 - 紧贴图片 */}
          {images.length < maxImages && (
            <Box
              sx={{
                border: '2px dashed',
                borderColor: isDragOver ? 'primary.main' : 'grey.300',
                borderRadius: 2,
                width: '80px',
                height: '80px',
                textAlign: 'center',
                cursor: 'pointer',
                bgcolor: isDragOver ? 'primary.50' : 'grey.100',
                transition: 'all 0.3s',
                position: 'relative',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
                '&:hover': {
                  borderColor: 'primary.light',
                  bgcolor: 'primary.50',
                },
              }}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              {isDragOver && (
                <Box
                  sx={{
                    position: 'absolute',
                    inset: 0,
                    bgcolor: 'primary.100',
                    border: '2px dashed',
                    borderColor: 'primary.main',
                    borderRadius: 2,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    zIndex: 1,
                  }}
                >
                  <DragIndicator sx={{ fontSize: 20, color: 'primary.dark' }} />
                </Box>
              )}
              
              <CloudUpload sx={{ fontSize: 24, color: 'primary.main' }} />
            </Box>
          )}

          {/* 图片数量显示和提示 */}
          <Box sx={{ 
            display: 'flex', 
            flexDirection: 'column',
            alignItems: 'flex-start',
            ml: 1,
            gap: 0.5,
          }}>
            {images.length >= maxImages ? (
              <Chip 
                label={`${images.length}/${maxImages} 已满`} 
                size="small" 
                color="success"
                sx={{ fontSize: '0.75rem' }}
              />
            ) : (
              <Chip 
                label={`${images.length}/${maxImages}`} 
                size="small" 
                color="primary"
                sx={{ fontSize: '0.75rem' }}
              />
            )}
            
            {images.length < maxImages && (
              <Typography 
                variant="caption" 
                color="text.secondary" 
                sx={{ fontSize: '0.65rem', lineHeight: 1.2 }}
              >
                💡 可按 Ctrl+V 粘贴
              </Typography>
            )}
          </Box>
        </Box>
      )}

      {/* 图片查看器 */}
      <ImageViewer
        open={viewerOpen}
        onClose={handleCloseViewer}
        imageUrl={images[currentImageIndex] || ''}
        title={`参考图片 ${currentImageIndex + 1}/${images.length}`}
      />
    </Paper>
  );
};

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
    const sizes = selectedModel?.image_sizes ?? [];
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
        size: selectedModel?.image_sizes?.length ? selectedSize : undefined,
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
                setSelectedSize(provider.models[0]?.image_sizes?.[0] ?? '');
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
                setSelectedSize(model.image_sizes?.[0] ?? '');
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
          {selectedModel?.image_sizes && selectedModel.image_sizes.length > 0 && (
            <FormControl fullWidth size="small">
              <InputLabel>图片尺寸</InputLabel>
              <Select
                value={selectedSize}
                label="图片尺寸"
                disabled={isGenerating}
                onChange={(e) => setSelectedSize(e.target.value)}
              >
                {selectedModel.image_sizes.map((sizeOption) => (
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
