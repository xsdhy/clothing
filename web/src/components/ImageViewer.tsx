import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Box, IconButton, Typography, Portal } from '@mui/material';
import { Close, Download, ZoomIn, ZoomOut, ZoomOutMap } from '@mui/icons-material';
import { PhotoSlider } from 'react-photo-view';
import type { DataType, OverlayRenderProps } from 'react-photo-view/dist/types';
import 'react-photo-view/dist/react-photo-view.css';

export interface ImageViewerProps {
  open: boolean;
  onClose: () => void;
  imageUrl?: string;
  images?: ImageViewerItem[];
  initialIndex?: number;
  title?: string;
  showDownload?: boolean;
  onDownload?: (index: number, item: ImageViewerItem) => void;
}

const MIN_SCALE = 0.2;
const MAX_SCALE = 5;

export interface ImageViewerItem {
  src: string;
  key?: string | number;
  title?: string;
  downloadName?: string;
}

type NormalizedImageViewerItem = ImageViewerItem & { key: string | number };

const ImageViewer: React.FC<ImageViewerProps> = ({
  open,
  onClose,
  imageUrl,
  images,
  initialIndex = 0,
  title,
  showDownload = false,
  onDownload,
}) => {
  const items = useMemo<NormalizedImageViewerItem[]>(() => {
    if (Array.isArray(images) && images.length > 0) {
      return images.map((item) => ({
        src: item.src,
        key: item.key ?? item.src,
        title: item.title,
        downloadName: item.downloadName,
      }));
    }
    if (!imageUrl) {
      return [];
    }
    return [{ src: imageUrl, key: imageUrl, title, downloadName: undefined }];
  }, [imageUrl, images, title]);

  const [currentIndex, setCurrentIndex] = useState(initialIndex);
  const [scale, setScale] = useState(1);
  const scaleCallbackRef = useRef<((scale: number) => void) | null>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    const safeIndex = Number.isFinite(initialIndex) ? Math.max(0, Math.min(initialIndex, items.length - 1)) : 0;
    setCurrentIndex(safeIndex);
    setScale(1);
  }, [initialIndex, items.length, open]);

  const currentItem = items[currentIndex];
  const resolvedTitle = currentItem?.title ?? title;
  const sliderImages = useMemo<DataType[]>(() => items.map(({ key, src }) => ({ key, src })), [items]);

  const handleZoomOut = () => {
    const newScale = Math.max(MIN_SCALE, scale / 1.2);
    setScale(newScale);
    scaleCallbackRef.current?.(newScale);
  };

  const handleReset = () => {
    setScale(1);
    scaleCallbackRef.current?.(1);
  };

  const handleZoomIn = () => {
    const newScale = Math.min(MAX_SCALE, scale * 1.2);
    setScale(newScale);
    scaleCallbackRef.current?.(newScale);
  };

  const handleDownload = () => {
    if (!onDownload || !currentItem) {
      return;
    }
    onDownload(currentIndex, currentItem);
  };

  const handleClose = () => {
    onClose();
  };

  const renderOverlay = ({ scale: photoScale, onScale }: OverlayRenderProps) => {
    // 保存 onScale 回调引用
    scaleCallbackRef.current = onScale || null;
    
    // 同步 PhotoSlider 的 scale 到我们的 state
    if (photoScale !== scale) {
      setScale(photoScale);
    }
    
    return null;
  };

  if (!open || items.length === 0) {
    return null;
  }

  return (
    <>
      <PhotoSlider
        visible={open}
        photoClosable={true}
        images={sliderImages}
        index={currentIndex}
        onClose={onClose}
        onIndexChange={(index) => {
          if (!Number.isFinite(index)) {
            return;
          }
          setCurrentIndex(index);
          setScale(1);
        }}
        overlayRender={renderOverlay}
        maskOpacity={0.95}
        bannerVisible={false}
        speed={() => 400}
        easing={() => 'cubic-bezier(0.4, 0, 0.2, 1)'}
      />
      
      {/* 独立的工具栏 Portal */}
      <Portal>
        <Box
          sx={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            zIndex: 2000,
            display: open ? 'grid' : 'none',
            gridTemplateColumns: { xs: '1fr', sm: 'minmax(0,1fr) auto' },
            alignItems: 'center',
            rowGap: 1,
            columnGap: 2,
            padding: 2,
            background: 'linear-gradient(180deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
            pointerEvents: 'none',
          }}
        >
          <Box sx={{ minWidth: 0, pointerEvents: 'auto' }}>
            {resolvedTitle && (
              <Typography
                variant="h6"
                sx={{ color: 'white', opacity: 0.9, fontSize: { xs: '1rem', sm: '1.1rem' } }}
                noWrap
              >
                {resolvedTitle}
              </Typography>
            )}
          </Box>

          <Box
            sx={{
              display: 'flex',
              gap: 1,
              pointerEvents: 'auto',
              justifyContent: { xs: 'flex-start', sm: 'flex-end' },
              flexWrap: 'wrap',
            }}
          >
            <IconButton
              onClick={handleZoomOut}
              sx={{
                color: 'white',
                bgcolor: 'rgba(255,255,255,0.12)',
                width: { xs: 40, sm: 42 },
                height: { xs: 40, sm: 42 },
                '&:hover': {
                  bgcolor: 'rgba(255,255,255,0.2)',
                },
              }}
              aria-label="zoom-out"
            >
              <ZoomOut />
            </IconButton>
            <IconButton
              onClick={handleReset}
              sx={{
                color: 'white',
                bgcolor: 'rgba(255,255,255,0.12)',
                width: { xs: 40, sm: 42 },
                height: { xs: 40, sm: 42 },
                '&:hover': {
                  bgcolor: 'rgba(255,255,255,0.2)',
                },
              }}
              aria-label="reset-zoom"
            >
              <ZoomOutMap />
            </IconButton>
            <IconButton
              onClick={handleZoomIn}
              sx={{
                color: 'white',
                bgcolor: 'rgba(255,255,255,0.12)',
                width: { xs: 40, sm: 42 },
                height: { xs: 40, sm: 42 },
                '&:hover': {
                  bgcolor: 'rgba(255,255,255,0.2)',
                },
              }}
              aria-label="zoom-in"
            >
              <ZoomIn />
            </IconButton>
            {showDownload && onDownload && (
              <IconButton
                onClick={handleDownload}
                sx={{
                  color: 'white',
                  bgcolor: 'rgba(255,255,255,0.12)',
                  width: { xs: 40, sm: 42 },
                  height: { xs: 40, sm: 42 },
                  '&:hover': {
                    bgcolor: 'rgba(255,255,255,0.2)',
                  },
                }}
                aria-label="download"
              >
                <Download />
              </IconButton>
            )}
            <IconButton
              onClick={handleClose}
              sx={{
                color: 'white',
                bgcolor: 'rgba(255,255,255,0.18)',
                width: { xs: 40, sm: 42 },
                height: { xs: 40, sm: 42 },
                '&:hover': {
                  bgcolor: 'rgba(255,255,255,0.28)',
                },
              }}
              aria-label="close"
            >
              <Close />
            </IconButton>
          </Box>
        </Box>

        {/* 底部信息栏 */}
        <Box
          sx={{
            position: 'fixed',
            bottom: 0,
            left: 0,
            right: 0,
            zIndex: 2000,
            display: open ? 'flex' : 'none',
            alignItems: 'center',
            justifyContent: 'center',
            px: { xs: 2, sm: 3 },
            py: { xs: 1.5, sm: 2 },
            pointerEvents: 'none',
            background: 'linear-gradient(0deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
          }}
        >
          <Typography
            variant="body2"
            sx={{
              color: 'white',
              opacity: 0.7,
              fontSize: { xs: '0.8rem', sm: '0.9rem' },
              textAlign: 'center',
            }}
          >
            缩放: {Math.round(scale * 100)}% | 滚轮缩放 | 拖拽移动
          </Typography>
        </Box>
      </Portal>
    </>
  );
};

export default ImageViewer;