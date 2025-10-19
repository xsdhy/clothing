import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  Box,
  Card,
  CardMedia,
  Chip,
  IconButton,
  Paper,
  Typography,
} from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import { CloudUpload, Delete, DragIndicator } from '@mui/icons-material';

import ImageViewer from './ImageViewer';

export interface ImageUploadProps {
  images: string[];
  onImagesChange: (images: string[]) => void;
  maxImages?: number;
  variant?: 'panel' | 'inline';
}

const ImageUpload: React.FC<ImageUploadProps> = ({
  images,
  onImagesChange,
  maxImages = 5,
  variant = 'panel',
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const [viewerOpen, setViewerOpen] = useState(false);
  const [currentImageIndex, setCurrentImageIndex] = useState(0);

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  const [isPasteActive, setIsPasteActive] = useState(false);
  const isInline = variant === 'inline';
  const theme = useTheme();

  const processFiles = useCallback(
    (files: File[]) => {
      if (files.length === 0) return;

      const MAX_FILE_SIZE = 5 * 1024 * 1024;

      if (images.length + files.length > maxImages) {
        alert(`最多只能上传 ${maxImages} 张图片`);
        return;
      }

      const readFileAsDataURL = (file: File): Promise<string> =>
        new Promise((resolve, reject) => {
          const reader = new FileReader();
          reader.onload = (e) => {
            resolve(e.target?.result as string);
          };
          reader.onerror = () => reject(new Error('文件读取失败'));
          reader.readAsDataURL(file);
        });

      const getBase64Size = (dataUrl: string): number => {
        const base64 = dataUrl.split(',')[1] || '';
        return Math.ceil((base64.length * 3) / 4);
      };

      const compressImage = async (file: File): Promise<string> => {
        const dataUrl = await readFileAsDataURL(file);
        if (file.size <= MAX_FILE_SIZE) {
          return dataUrl;
        }

        return new Promise((resolve, reject) => {
          const img = new Image();
          img.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            if (!ctx) {
              reject(new Error('浏览器不支持图片压缩'));
              return;
            }

            let targetWidth = img.width;
            let targetHeight = img.height;
            let quality = 0.9;
            const outputType = file.type === 'image/png' ? 'image/png' : 'image/jpeg';
            let compressedDataUrl = dataUrl;
            let attempts = 0;

            const ratio = Math.sqrt(MAX_FILE_SIZE / file.size);
            if (ratio < 1) {
              targetWidth = Math.max(Math.floor(img.width * ratio), 1);
              targetHeight = Math.max(Math.floor(img.height * ratio), 1);
            }

            const compressLoop = () => {
              canvas.width = targetWidth;
              canvas.height = targetHeight;
              ctx.clearRect(0, 0, targetWidth, targetHeight);
              ctx.drawImage(img, 0, 0, targetWidth, targetHeight);
              compressedDataUrl = canvas.toDataURL(outputType, quality);

              const size = getBase64Size(compressedDataUrl);
              if (size <= MAX_FILE_SIZE) {
                resolve(compressedDataUrl);
                return;
              }

              if (attempts >= 20) {
                reject(new Error('图片压缩失败，请选择更小的图片'));
                return;
              }

              attempts += 1;

              if (outputType === 'image/jpeg' && quality > 0.5) {
                quality = Math.max(quality - 0.1, 0.5);
              } else {
                targetWidth = Math.max(Math.floor(targetWidth * 0.85), 1);
                targetHeight = Math.max(Math.floor(targetHeight * 0.85), 1);
              }

              requestAnimationFrame(compressLoop);
            };

            compressLoop();
          };
          img.onerror = () => reject(new Error('图片加载失败，无法压缩'));
          img.src = dataUrl;
        });
      };

      Promise.all(
        files.map((file) => {
          if (!file.type.startsWith('image/')) {
            return Promise.reject(new Error('请选择图片文件'));
          }

          return compressImage(file);
        })
      )
        .then((base64Images) => {
          onImagesChange([...images, ...base64Images]);
        })
        .catch((error) => {
          alert(error.message);
        });
    },
    [images, onImagesChange, maxImages]
  );

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || []);
    processFiles(files);

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

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

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragOver(false);

      const files = Array.from(e.dataTransfer.files);
      processFiles(files);
    },
    [processFiles]
  );

  const removeImage = (index: number) => {
    const newImages = [...images];
    newImages.splice(index, 1);
    onImagesChange(newImages);
  };

  const handleImageClick = (index: number) => {
    setCurrentImageIndex(index);
    setViewerOpen(true);
  };

  const handleCloseViewer = () => {
    setViewerOpen(false);
  };

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

      newImages.splice(draggedIndex, 1);
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

  const handlePaste = useCallback(
    (e: ClipboardEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
        return;
      }

      if (!isPasteActive) {
        return;
      }

      const items = Array.from(e.clipboardData?.items || []);
      const imageItems = items.filter((item) => item.type.startsWith('image/'));
      if (imageItems.length === 0) {
        return;
      }

      const files = imageItems
        .map((item) => item.getAsFile())
        .filter((file): file is File => Boolean(file));

      if (files.length > 0) {
        e.preventDefault();
        processFiles(files);
      }
    },
    [isPasteActive, processFiles]
  );

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'v') {
        setIsPasteActive(true);
      }
    };

    const handleKeyUp = (event: KeyboardEvent) => {
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'v') {
        setIsPasteActive(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    window.addEventListener('paste', handlePaste as EventListener);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
      window.removeEventListener('paste', handlePaste as EventListener);
    };
  }, [handlePaste]);

  const panelThumbnailSize = 80;
  const inlineThumbnailSize = 96;

  const renderThumbnail = (image: string, index: number, size: number) => {
    const isDragged = draggedIndex === index;
    const isDropTarget = dragOverIndex === index;

    return (
      <Card
        key={`${image}-${index}`}
        draggable
        onDragStart={(e) => handleImageDragStart(e, index)}
        onDragOver={(e) => handleImageDragOver(e, index)}
        onDragLeave={handleImageDragLeave}
        onDrop={(e) => handleImageDrop(e, index)}
        onDragEnd={handleImageDragEnd}
        sx={{
          position: 'relative',
          width: size,
          height: size,
          cursor: isDragged ? 'grabbing' : 'grab',
          opacity: isDragged ? 0.55 : 1,
          transform: isDragged ? 'scale(0.96)' : isDropTarget ? 'scale(1.04)' : 'scale(1)',
          borderRadius: 3,
          overflow: 'hidden',
          flexShrink: 0,
          border: '1px solid',
          borderColor: isDropTarget ? theme.palette.primary.main : alpha('#0F172A', 0.08),
          transition: 'all 0.24s ease',
          boxShadow: isDropTarget
            ? '0 18px 38px rgba(99,102,241,0.25)'
            : '0 14px 32px rgba(15,23,42,0.12)',
          background: 'linear-gradient(135deg, rgba(99,102,241,0.08) 0%, rgba(255,255,255,1) 100%)',
          '&:hover .delete-btn': {
            opacity: 1,
          },
          '&:hover .drag-indicator': {
            opacity: 1,
          },
          '&:hover': {
            transform: isDragged ? 'scale(0.96)' : 'scale(1.05)',
            boxShadow: '0 24px 45px rgba(15,23,42,0.18)',
          },
        }}
        onClick={() => !isDragged && handleImageClick(index)}
      >
        <CardMedia
          component="img"
          image={image}
          alt={`参考图片 ${index + 1}`}
          sx={{ width: '100%', height: '100%', objectFit: 'cover' }}
        />
        <Box
          className="drag-indicator"
          sx={{
            position: 'absolute',
            top: 6,
            left: 6,
            bgcolor: alpha('#0F172A', 0.68),
            color: '#FFFFFF',
            borderRadius: 1,
            width: 24,
            height: 24,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            opacity: isDragged ? 1 : 0,
            transition: 'opacity 0.2s',
            boxShadow: '0 8px 14px rgba(15,23,42,0.25)',
          }}
        >
          <DragIndicator sx={{ fontSize: 15 }} />
        </Box>

        <IconButton
          className="delete-btn"
          size="small"
          onClick={(e) => {
            e.stopPropagation();
            removeImage(index);
          }}
          sx={{
            position: 'absolute',
            top: 6,
            right: 6,
            bgcolor: alpha(theme.palette.error.main, 0.92),
            color: '#FFFFFF',
            opacity: 0,
            transition: 'opacity 0.2s',
            width: 24,
            height: 24,
            '&:hover': {
              bgcolor: alpha(theme.palette.error.main, 1),
            },
            boxShadow: '0 10px 18px rgba(239,68,68,0.25)',
          }}
        >
          <Delete sx={{ fontSize: 16 }} />
        </IconButton>
        <Box
          sx={{
            position: 'absolute',
            bottom: 6,
            left: 6,
            bgcolor: alpha('#0F172A', 0.72),
            color: '#FFFFFF',
            px: 1,
            py: 0.25,
            borderRadius: 1,
            fontSize: '0.7rem',
            lineHeight: 1,
            letterSpacing: 0.4,
            boxShadow: '0 8px 16px rgba(15,23,42,0.25)',
          }}
        >
          #{index + 1}
        </Box>
      </Card>
    );
  };

  const panelContent = (
    <Paper
      variant="outlined"
      sx={{
        p: { xs: 3, md: 3.5 },
        borderRadius: 3,
        borderColor: alpha(theme.palette.primary.main, 0.08),
        background: 'linear-gradient(180deg, rgba(255,255,255,1) 0%, rgba(244,246,255,0.88) 100%)',
        boxShadow: '0 24px 60px rgba(99,102,241,0.12)',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          alignItems: { xs: 'flex-start', sm: 'center' },
          justifyContent: 'space-between',
          flexDirection: { xs: 'column', sm: 'row' },
          gap: 1.5,
          mb: 3,
        }}
      >
        <Box>
          <Typography variant="h6" sx={{ fontWeight: 700, color: 'text.primary' }}>
            参考图片上传
          </Typography>
          <Typography variant="body2" color="text.secondary">
            上传参考图可帮助模型理解构图、色彩与细节
          </Typography>
        </Box>
        <Chip
          label="支持拖拽 / 点击 / 粘贴"
          variant="outlined"
          sx={{
            borderColor: 'transparent',
            backgroundColor: alpha(theme.palette.primary.main, 0.12),
            color: theme.palette.primary.main,
            fontWeight: 600,
            borderRadius: '999px',
            px: 1.5,
            height: 28,
          }}
        />
      </Box>

      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        multiple
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />

      <Box
        sx={{
          position: 'relative',
          borderRadius: 3,
          border: '1.5px dashed',
          borderColor: isDragOver ? theme.palette.primary.main : alpha(theme.palette.primary.main, 0.24),
          background: isDragOver
            ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.16)} 0%, ${alpha(theme.palette.primary.main, 0.04)} 100%)`
            : 'linear-gradient(135deg, rgba(99,102,241,0.05) 0%, rgba(255,255,255,1) 100%)',
          minHeight: { xs: 180, md: 200 },
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          textAlign: 'center',
          p: { xs: 3, md: 4 },
          mb: 3,
          cursor: 'pointer',
          boxShadow: isDragOver ? '0 32px 68px rgba(99,102,241,0.24)' : 'inset 0 0 0 1px rgba(255,255,255,0.6)',
          overflow: 'hidden',
          transition: 'all 0.35s ease',
        }}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
      >
        <Box
          sx={{
            position: 'absolute',
            inset: 0,
            pointerEvents: 'none',
            opacity: isDragOver ? 1 : 0,
            transition: 'opacity 0.25s ease',
            background: `radial-gradient(circle at center, ${alpha(theme.palette.primary.main, 0.2)} 0%, transparent 70%)`,
          }}
        />
        <Box
          sx={{
            position: 'relative',
            zIndex: 1,
            maxWidth: 360,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 1.5,
          }}
        >
          <Box
            sx={{
              width: 68,
              height: 68,
              borderRadius: 2.5,
              background: 'linear-gradient(135deg, rgba(99,102,241,1) 0%, rgba(129,140,248,1) 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#FFFFFF',
              boxShadow: '0 24px 45px rgba(99,102,241,0.32)',
            }}
          >
            <CloudUpload sx={{ fontSize: 32 }} />
          </Box>
          <Box>
            <Typography variant="h6" component="div" sx={{ fontWeight: 700, mb: 1 }}>
              将图片拖拽到这里，或点击上传
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
              支持 JPG / PNG / WebP，单张图片不超过 5MB
            </Typography>
            <Typography variant="caption" color="text.secondary">
              Tip: 使用 Ctrl / Cmd + V 直接粘贴剪贴板图片
            </Typography>
          </Box>
        </Box>
      </Box>

      {images.length > 0 && (
        <Box
          sx={{
            display: 'flex',
            gap: 1.5,
            flexWrap: 'nowrap',
            overflowX: 'auto',
            pb: 1,
            px: 0.5,
          }}
        >
          {images.map((image, index) => (
            <Box key={`${image}-${index}`} sx={{ flexShrink: 0 }}>
              {renderThumbnail(image, index, panelThumbnailSize)}
            </Box>
          ))}

          {images.length < maxImages && (
            <Box
              sx={{
                width: panelThumbnailSize,
                height: panelThumbnailSize,
                borderRadius: 3,
                border: '1.5px dashed',
                borderColor: isDragOver ? theme.palette.primary.main : alpha(theme.palette.primary.main, 0.28),
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: theme.palette.primary.main,
                cursor: 'pointer',
                flexShrink: 0,
                background: 'linear-gradient(135deg, rgba(99,102,241,0.05) 0%, rgba(255,255,255,1) 100%)',
                transition: 'all 0.25s ease',
                boxShadow: '0 16px 32px rgba(99,102,241,0.15)',
                '&:hover': {
                  borderColor: theme.palette.primary.main,
                  background: 'linear-gradient(135deg, rgba(99,102,241,0.12) 0%, rgba(255,255,255,1) 100%)',
                  transform: 'translateY(-2px)',
                },
              }}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              <CloudUpload sx={{ fontSize: 28 }} />
            </Box>
          )}

          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              gap: 0.75,
              minWidth: 120,
            }}
          >
            {images.length >= maxImages ? (
              <Chip
                label={`${images.length}/${maxImages} 已满`}
                sx={{
                  borderColor: 'transparent',
                  backgroundColor: alpha(theme.palette.success.main, 0.2),
                  color: theme.palette.success.dark,
                  fontWeight: 600,
                  borderRadius: '999px',
                  height: 28,
                }}
              />
            ) : (
              <Chip
                label={`${images.length}/${maxImages}`}
                sx={{
                  borderColor: 'transparent',
                  backgroundColor: alpha(theme.palette.primary.main, 0.12),
                  color: theme.palette.primary.main,
                  fontWeight: 600,
                  borderRadius: '999px',
                  height: 28,
                }}
              />
            )}

            {images.length < maxImages && (
              <Typography variant="caption" color="text.secondary">
                💡 可再添加 {maxImages - images.length} 张参考图片
              </Typography>
            )}
          </Box>
        </Box>
      )}
    </Paper>
  );

  const inlineContent = (
    <Box
      sx={{
        display: 'flex',
        gap: { xs: 1.5, sm: 2 },
        alignItems: { xs: 'stretch', sm: 'flex-start' },
        width: '100%',
      }}
    >
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        multiple
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />
      <Box
        sx={{
          borderRadius: 3,
          border: '1.5px dashed',
          borderColor: isDragOver ? theme.palette.primary.main : alpha(theme.palette.primary.main, 0.24),
          background: isDragOver
            ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.18)} 0%, ${alpha(theme.palette.primary.main, 0.06)} 100%)`
            : 'linear-gradient(135deg, rgba(99,102,241,0.04) 0%, rgba(255,255,255,0.85) 100%)',
          display: 'flex',
          flexDirection: { xs: 'row', sm: 'column' },
          gap: 1,
          p: 1,
          width: { xs: '100%', sm: inlineThumbnailSize + 28 },
          minHeight: inlineThumbnailSize + 28,
          overflowX: { xs: 'auto', sm: 'hidden' },
          overflowY: { xs: 'hidden', sm: 'auto' },
          boxShadow: isDragOver ? '0 24px 55px rgba(99,102,241,0.2)' : 'inset 0 0 0 1px rgba(255,255,255,0.5)',
          transition: 'all 0.3s ease',
        }}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        {images.map((image, index) => renderThumbnail(image, index, inlineThumbnailSize))}

        {images.length < maxImages && (
          <Box
            sx={{
              width: inlineThumbnailSize,
              height: inlineThumbnailSize,
              borderRadius: 3,
              border: '1.5px dashed',
              borderColor: isDragOver ? theme.palette.primary.main : alpha(theme.palette.primary.main, 0.28),
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: theme.palette.primary.main,
              cursor: 'pointer',
              flexShrink: 0,
              background: 'linear-gradient(135deg, rgba(99,102,241,0.05) 0%, rgba(255,255,255,1) 100%)',
              transition: 'all 0.25s ease',
              boxShadow: '0 14px 28px rgba(99,102,241,0.12)',
              '&:hover': {
                borderColor: theme.palette.primary.main,
                background: 'linear-gradient(135deg, rgba(99,102,241,0.12) 0%, rgba(255,255,255,1) 100%)',
                transform: 'translateY(-2px)',
              },
            }}
            onClick={(event) => {
              event.stopPropagation();
              fileInputRef.current?.click();
            }}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            <CloudUpload sx={{ fontSize: 26 }} />
          </Box>
        )}
      </Box>
    </Box>
  );

  return (
    <>
      {isInline ? inlineContent : panelContent}
      <ImageViewer
        open={viewerOpen}
        onClose={handleCloseViewer}
        imageUrl={images[currentImageIndex] || ''}
        title={`参考图片 ${currentImageIndex + 1}/${images.length}`}
      />
    </>
  );
};

export default ImageUpload;
