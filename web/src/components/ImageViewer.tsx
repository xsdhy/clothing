import React, { useEffect, useRef, useState } from 'react';
import {
  Box,
  Dialog,
  DialogContent,
  IconButton,
  Typography,
  Zoom,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { Close, Download, ZoomIn, ZoomOut, ZoomOutMap } from '@mui/icons-material';

export interface ImageViewerProps {
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

  const resetView = () => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
  };

  const handleClose = () => {
    resetView();
    onClose();
  };

  const handleZoomIn = () => {
    setScale((prev) => Math.min(prev * 1.2, 5));
  };

  const handleZoomOut = () => {
    setScale((prev) => Math.max(prev / 1.2, 0.1));
  };

  const handleResetZoom = () => {
    resetView();
  };

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

  useEffect(() => {
    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      const delta = e.deltaY > 0 ? 0.9 : 1.1;
      setScale((prev) => Math.max(0.1, Math.min(5, prev * delta)));
    };

    const dialogContent = dialogContentRef.current;
    if (dialogContent && open) {
      dialogContent.addEventListener('wheel', handleWheel, { passive: false });
      return () => {
        dialogContent.removeEventListener('wheel', handleWheel);
      };
    }

    return undefined;
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
        <Box sx={{ flex: 1 }}>
          {title && (
            <Typography variant="h6" sx={{ color: 'white', opacity: 0.9 }}>
              {title}
            </Typography>
          )}
        </Box>

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

export default ImageViewer;
