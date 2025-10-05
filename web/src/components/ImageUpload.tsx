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
import { CloudUpload, Delete, DragIndicator } from '@mui/icons-material';

import ImageViewer from './ImageViewer';

export interface ImageUploadProps {
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

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  const [isPasteActive, setIsPasteActive] = useState(false);

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

  return (
    <Paper sx={{ p: 2, borderRadius: 2 }} variant="outlined">
      <Typography variant="h6" sx={{ mb: 2 }}>
        参考图片上传
      </Typography>

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
          border: '2px dashed',
          borderColor: isDragOver ? 'primary.main' : 'grey.300',
          borderRadius: 2,
          height: 160,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          textAlign: 'center',
          p: 2,
          mb: 2,
          position: 'relative',
          overflow: 'hidden',
          bgcolor: isDragOver ? 'primary.50' : 'background.default',
          transition: 'all 0.3s',
          cursor: 'pointer',
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
              bgcolor: 'rgba(63,81,181,0.1)',
              border: '2px dashed',
              borderColor: 'primary.main',
              borderRadius: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              zIndex: 1,
            }}
          >
            <Typography variant="subtitle1" color="primary">
              松开鼠标上传图片
            </Typography>
          </Box>
        )}

        <Box sx={{ position: 'relative', zIndex: 2 }}>
          <CloudUpload sx={{ fontSize: 36, color: 'primary.main', mb: 1 }} />
          <Typography variant="h6" component="div" gutterBottom>
            拖拽图片到这里，或点击上传
          </Typography>
          <Typography variant="body2" color="text.secondary">
            支持 JPG / PNG / WebP，单张图片不超过 5MB
          </Typography>
          <Typography variant="caption" color="text.secondary">
            可按 CTRL+V 粘贴剪贴板中的图片
          </Typography>
        </Box>
      </Box>

      {images.length > 0 && (
        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'nowrap', overflowX: 'auto', pb: 1 }}>
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

          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-start',
              ml: 1,
              gap: 0.5,
            }}
          >
            {images.length >= maxImages ? (
              <Chip label={`${images.length}/${maxImages} 已满`} size="small" color="success" sx={{ fontSize: '0.75rem' }} />
            ) : (
              <Chip label={`${images.length}/${maxImages}`} size="small" color="primary" sx={{ fontSize: '0.75rem' }} />
            )}

            {images.length < maxImages && (
              <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.65rem', lineHeight: 1.2 }}>
                💡 可按 Ctrl+V 粘贴
              </Typography>
            )}
          </Box>
        </Box>
      )}

      <ImageViewer
        open={viewerOpen}
        onClose={handleCloseViewer}
        imageUrl={images[currentImageIndex] || ''}
        title={`参考图片 ${currentImageIndex + 1}/${images.length}`}
      />
    </Paper>
  );
};

export default ImageUpload;
