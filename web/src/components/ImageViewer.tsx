import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Box, IconButton, Typography, Portal } from '@mui/material';
import { Close, Download, ZoomIn, ZoomOut, ZoomOutMap } from '@mui/icons-material';
import { PhotoSlider } from 'react-photo-view';
import type { DataType, OverlayRenderProps } from 'react-photo-view/dist/types';
import 'react-photo-view/dist/react-photo-view.css';
import ReactPlayer from 'react-player';

import { isVideoUrl } from '../utils/media';

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
  type?: 'image' | 'video';
}

type NormalizedImageViewerItem = ImageViewerItem & { key: string | number };

const BLANK_PLACEHOLDER = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==';

const isVideoItem = (item?: ImageViewerItem): boolean => {
  if (!item) {
    return false;
  }
  if (item.type === 'video') {
    return true;
  }
  return isVideoUrl(item.src);
};

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
        type: item.type ?? (isVideoUrl(item.src) ? 'video' : 'image'),
      }));
    }
    if (!imageUrl) {
      return [];
    }
    return [
      {
        src: imageUrl,
        key: imageUrl,
        title,
        downloadName: undefined,
        type: isVideoUrl(imageUrl) ? 'video' : 'image',
      },
    ];
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
  const sliderImages = useMemo<DataType[]>(
    () =>
      items.map(({ key, src, type }) => ({
        key,
        src: type === 'video' ? BLANK_PLACEHOLDER : src,
      })),
    [items],
  );
  const currentIsVideo = isVideoItem(currentItem);

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
    if (currentIsVideo) {
      scaleCallbackRef.current = null;
      return (
        <Box
          sx={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            pointerEvents: 'none',
            px: { xs: 1.5, sm: 3 },
          }}
        >
          <Box
            sx={{
              position: 'relative',
              width: 'min(100%, 1280px)',
              maxHeight: 'calc(100vh - 160px)',
              aspectRatio: '16 / 9',
              boxShadow: 8,
              borderRadius: 2,
              bgcolor: 'black',
              border: '1px solid rgba(255,255,255,0.08)',
              overflow: 'hidden',
              pointerEvents: 'auto',
            }}
          >
            <ReactPlayer
              src={currentItem?.src}
              playing
              loop
              controls
              width="100%"
              height="100%"
              style={{ position: 'absolute', inset: 0 }}
              playsInline
              config={{ html: { controlsList: 'nodownload' } }}
            />
          </Box>
        </Box>
      );
    }

    // 保存 onScale 回调引用（仅图片需要缩放）
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
            {!currentIsVideo && (
              <>
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
              </>
            )}
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
            {currentIsVideo ? '视频播放' : `缩放: ${Math.round(scale * 100)}%`}
          </Typography>
        </Box>
      </Portal>
    </>
  );
};

export default ImageViewer;
