import React, { useMemo } from 'react';
import { Box, IconButton, Typography } from '@mui/material';
import { Close, Download, ZoomIn, ZoomOut, ZoomOutMap } from '@mui/icons-material';
import { PhotoSlider } from 'react-photo-view';
import type { OverlayRenderProps } from 'react-photo-view/dist/types';
import 'react-photo-view/dist/react-photo-view.css';

export interface ImageViewerProps {
  open: boolean;
  onClose: () => void;
  imageUrl: string;
  title?: string;
  showDownload?: boolean;
  onDownload?: () => void;
}

const MIN_SCALE = 0.2;
const MAX_SCALE = 5;

const ImageViewer: React.FC<ImageViewerProps> = ({
  open,
  onClose,
  imageUrl,
  title,
  showDownload = false,
  onDownload,
}) => {
  const images = useMemo(() => {
    if (!imageUrl) {
      return [];
    }
    return [{ src: imageUrl, key: imageUrl }];
  }, [imageUrl]);

  const handleToolbarRender = ({ scale = 1, onScale, onClose: sliderClose }: OverlayRenderProps) => {
    const applyScale = (next: number) => {
      if (!onScale) {
        return;
      }
      const clamped = Math.max(MIN_SCALE, Math.min(MAX_SCALE, next));
      onScale(clamped);
    };

    const handleZoomOut = (event: React.MouseEvent) => {
      event.stopPropagation();
      applyScale(scale / 1.2);
    };

    const handleReset = (event: React.MouseEvent) => {
      event.stopPropagation();
      applyScale(1);
    };

    const handleZoomIn = (event: React.MouseEvent) => {
      event.stopPropagation();
      applyScale(scale * 1.2);
    };

    const handleDownload = (event: React.MouseEvent) => {
      event.stopPropagation();
      onDownload?.();
    };

    const handleClose = (event: React.MouseEvent) => {
      event.stopPropagation();
      sliderClose?.(event);
    };

    return (
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          p: 2,
          gap: 2,
          pointerEvents: 'none',
          background: 'linear-gradient(180deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
        }}
      >
        <Box sx={{ flex: 1, minWidth: 0, pointerEvents: 'auto' }}>
          {title && (
            <Typography variant="h6" sx={{ color: 'white', opacity: 0.9 }} noWrap>
              {title}
            </Typography>
          )}
        </Box>

        <Box sx={{ display: 'flex', gap: 1, pointerEvents: 'auto' }}>
          <IconButton onClick={handleZoomOut} sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }} size="small">
            <ZoomOut />
          </IconButton>
          <IconButton onClick={handleReset} sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }} size="small">
            <ZoomOutMap />
          </IconButton>
          <IconButton onClick={handleZoomIn} sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }} size="small">
            <ZoomIn />
          </IconButton>
          {showDownload && onDownload && (
            <IconButton onClick={handleDownload} sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }} size="small">
              <Download />
            </IconButton>
          )}
          <IconButton onClick={handleClose} sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }} size="small">
            <Close />
          </IconButton>
        </Box>
      </Box>
    );
  };

  const handleOverlayRender = ({ scale = 1 }: OverlayRenderProps) => (
    <Box
      sx={{
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        p: 2,
        pointerEvents: 'none',
        background: 'linear-gradient(0deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
      }}
    >
      <Typography variant="body2" sx={{ color: 'white', opacity: 0.7 }}>
        缩放: {Math.round(scale * 100)}% | 滚轮缩放 | 拖拽移动
      </Typography>
    </Box>
  );

  if (!open || images.length === 0) {
    return null;
  }

  return (
    <PhotoSlider
      visible={open}
      images={images}
      index={0}
      onClose={() => {
        onClose();
      }}
      toolbarRender={handleToolbarRender}
      overlayRender={handleOverlayRender}
      maskOpacity={0.95}
    />
  );
};

export default ImageViewer;
