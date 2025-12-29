import React, { useEffect, useRef, useState } from "react";
import {
  Box,
  IconButton,
  LinearProgress,
  Typography,
  useMediaQuery,
} from "@mui/material";
import type { SxProps, Theme } from "@mui/material/styles";
import {
  OpenInFullRounded,
  PauseRounded,
  PlayArrowRounded,
  VolumeOffRounded,
  VolumeUpRounded,
} from "@mui/icons-material";

export interface VideoPlayerProps {
  src: string;
  title?: string;
  poster?: string;
  aspectRatio?: string | number;
  fit?: "cover" | "contain";
  loop?: boolean;
  autoPlay?: boolean;
  muted?: boolean;
  compact?: boolean;
  showControls?: boolean;
  showTimeline?: boolean;
  allowToggleOnClick?: boolean;
  onOpen?: () => void;
  openLabel?: string;
  coverLabel?: string;
  sx?: SxProps<Theme>;
}

const formatTime = (value: number): string => {
  if (!Number.isFinite(value) || value < 0) {
    return "0:00";
  }
  const total = Math.floor(value);
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const seconds = total % 60;
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
};

const VideoPlayer: React.FC<VideoPlayerProps> = ({
  src,
  title,
  poster,
  aspectRatio,
  fit = "cover",
  loop = true,
  autoPlay = false,
  muted,
  compact = false,
  showControls = true,
  showTimeline,
  allowToggleOnClick = true,
  onOpen,
  openLabel = "打开预览",
  coverLabel = "点击播放",
  sx,
}) => {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isMuted, setIsMuted] = useState(muted ?? autoPlay);
  const [duration, setDuration] = useState(0);
  const [currentTime, setCurrentTime] = useState(0);
  const [hovered, setHovered] = useState(false);
  const canHover = useMediaQuery("(hover: hover)");

  useEffect(() => {
    setIsPlaying(false);
    setDuration(0);
    setCurrentTime(0);
  }, [src]);

  useEffect(() => {
    if (muted === undefined) {
      return;
    }
    setIsMuted(muted);
  }, [muted]);

  useEffect(() => {
    if (videoRef.current) {
      videoRef.current.muted = isMuted;
    }
  }, [isMuted]);

  useEffect(() => {
    if (!autoPlay) {
      return;
    }
    const video = videoRef.current;
    if (!video) {
      return;
    }
    const attemptPlay = async () => {
      try {
        await video.play();
      } catch {
        // Autoplay may be blocked; ignore and wait for user interaction.
      }
    };
    void attemptPlay();
  }, [autoPlay, src]);

  const resolvedShowTimeline = showTimeline ?? !compact;
  const showControlsOverlay =
    showControls && isPlaying && (!canHover || hovered);

  const handleTogglePlay = () => {
    const video = videoRef.current;
    if (!video) {
      return;
    }
    if (video.paused) {
      const result = video.play();
      if (result && typeof result.catch === "function") {
        void result.catch(() => {});
      }
    } else {
      video.pause();
    }
  };

  const handleToggleMute = () => {
    setIsMuted((prev) => !prev);
  };

  const handleSeek = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!duration) {
      return;
    }
    const rect = event.currentTarget.getBoundingClientRect();
    const ratio = Math.min(
      1,
      Math.max(0, (event.clientX - rect.left) / rect.width),
    );
    const video = videoRef.current;
    if (!video) {
      return;
    }
    video.currentTime = ratio * duration;
    setCurrentTime(video.currentTime);
  };

  const controlSize = compact ? 30 : 38;
  const iconFontSize = compact ? "small" : "medium";
  const showTitle = Boolean(title) && !compact;

  const baseSx: SxProps<Theme> = {
    position: "relative",
    borderRadius: 2,
    overflow: "hidden",
    bgcolor: "rgba(7, 11, 18, 0.9)",
  };

  const resolvedSxList = Array.isArray(sx) ? sx : sx ? [sx] : [];
  const mergedSx: SxProps<Theme> = [
    baseSx,
    aspectRatio ? { aspectRatio } : {},
    ...resolvedSxList,
  ];

  return (
    <Box
      sx={mergedSx}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <Box
        component="video"
        ref={videoRef}
        src={src}
        poster={poster}
        loop={loop}
        autoPlay={autoPlay}
        muted={isMuted}
        playsInline
        preload="metadata"
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        onEnded={() => setIsPlaying(false)}
        onLoadedMetadata={(event) => {
          setDuration(event.currentTarget.duration || 0);
        }}
        onTimeUpdate={(event) => {
          setCurrentTime(event.currentTarget.currentTime || 0);
        }}
        onClick={(event) => {
          if (!allowToggleOnClick) {
            return;
          }
          event.stopPropagation();
          handleTogglePlay();
        }}
        className="video-player-media"
        sx={{
          width: "100%",
          height: "100%",
          display: "block",
          objectFit: fit,
          backgroundColor: "black",
        }}
      />

      {!isPlaying && (
        <Box
          sx={{
            position: "absolute",
            inset: 0,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            gap: compact ? 1 : 1.5,
            background:
              "linear-gradient(180deg, rgba(8, 12, 20, 0.6) 0%, rgba(8, 12, 20, 0.85) 100%)",
            color: "common.white",
            textAlign: "center",
            cursor: "pointer",
          }}
          onClick={(event) => {
            event.stopPropagation();
            handleTogglePlay();
          }}
        >
          <Box
            sx={{
              width: compact ? 52 : 70,
              height: compact ? 52 : 70,
              borderRadius: "999px",
              background:
                "radial-gradient(circle at 30% 30%, rgba(255,255,255,0.28), rgba(255,255,255,0.08) 45%, rgba(0,0,0,0.4) 100%)",
              border: "1px solid rgba(255,255,255,0.28)",
              display: "grid",
              placeItems: "center",
              boxShadow: "0 18px 36px rgba(0,0,0,0.35)",
              position: "relative",
              "&::before": {
                content: '""',
                position: "absolute",
                inset: -6,
                borderRadius: "999px",
                border: "1px solid rgba(255,255,255,0.2)",
                opacity: 0.6,
                animation: "video-play-pulse 2.4s ease-in-out infinite",
              },
              "@keyframes video-play-pulse": {
                "0%": { transform: "scale(0.98)", opacity: 0.6 },
                "50%": { transform: "scale(1.08)", opacity: 0.3 },
                "100%": { transform: "scale(0.98)", opacity: 0.6 },
              },
            }}
          >
            <PlayArrowRounded sx={{ fontSize: compact ? 30 : 36, ml: 0.5 }} />
          </Box>
          <Typography
            variant="caption"
            sx={{
              letterSpacing: "0.12em",
              textTransform: "uppercase",
              opacity: 0.9,
              fontSize: compact ? "0.65rem" : "0.75rem",
            }}
          >
            {coverLabel}
          </Typography>
          {onOpen && (
            <IconButton
              size={compact ? "small" : "medium"}
              aria-label={openLabel}
              onClick={(event) => {
                event.stopPropagation();
                onOpen();
              }}
              sx={{
                position: "absolute",
                top: compact ? 6 : 10,
                right: compact ? 6 : 10,
                bgcolor: "rgba(0,0,0,0.5)",
                color: "common.white",
                "&:hover": {
                  bgcolor: "rgba(0,0,0,0.7)",
                },
              }}
            >
              <OpenInFullRounded fontSize={iconFontSize} />
            </IconButton>
          )}
        </Box>
      )}

      {showControlsOverlay && (
        <Box
          sx={{
            position: "absolute",
            inset: 0,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
            pointerEvents: "none",
          }}
        >
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              px: compact ? 1 : 1.5,
              py: compact ? 0.75 : 1,
              pointerEvents: "auto",
              background:
                "linear-gradient(180deg, rgba(0,0,0,0.6) 0%, rgba(0,0,0,0) 100%)",
            }}
          >
            {showTitle && (
              <Typography
                variant="subtitle2"
                sx={{
                  color: "common.white",
                  textShadow: "0 2px 10px rgba(0,0,0,0.35)",
                }}
                noWrap
              >
                {title}
              </Typography>
            )}
            {onOpen && (
              <IconButton
                size={compact ? "small" : "medium"}
                aria-label={openLabel}
                onClick={(event) => {
                  event.stopPropagation();
                  onOpen();
                }}
                sx={{
                  bgcolor: "rgba(0,0,0,0.45)",
                  color: "common.white",
                  "&:hover": {
                    bgcolor: "rgba(0,0,0,0.65)",
                  },
                }}
              >
                <OpenInFullRounded fontSize={iconFontSize} />
              </IconButton>
            )}
          </Box>
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              gap: 1,
              px: compact ? 1 : 1.5,
              py: compact ? 0.75 : 1,
              pointerEvents: "auto",
              background:
                "linear-gradient(0deg, rgba(0,0,0,0.7) 0%, rgba(0,0,0,0) 100%)",
            }}
          >
            <IconButton
              size={compact ? "small" : "medium"}
              aria-label={isPlaying ? "暂停" : "播放"}
              onClick={(event) => {
                event.stopPropagation();
                handleTogglePlay();
              }}
              sx={{
                bgcolor: "rgba(255,255,255,0.12)",
                color: "common.white",
                "&:hover": {
                  bgcolor: "rgba(255,255,255,0.22)",
                },
              }}
            >
              {isPlaying ? (
                <PauseRounded fontSize={iconFontSize} />
              ) : (
                <PlayArrowRounded fontSize={iconFontSize} />
              )}
            </IconButton>

            {resolvedShowTimeline && (
              <Box
                onClick={handleSeek}
                sx={{
                  flex: 1,
                  cursor: "pointer",
                  display: "flex",
                  alignItems: "center",
                  height: compact ? 16 : 20,
                }}
              >
                <LinearProgress
                  variant="determinate"
                  value={duration ? (currentTime / duration) * 100 : 0}
                  sx={{
                    width: "100%",
                    height: compact ? 4 : 6,
                    borderRadius: 999,
                    bgcolor: "rgba(255,255,255,0.2)",
                    "& .MuiLinearProgress-bar": {
                      borderRadius: 999,
                      bgcolor: "rgba(255,255,255,0.85)",
                    },
                  }}
                />
              </Box>
            )}

            {resolvedShowTimeline && (
              <Typography
                variant="caption"
                sx={{ color: "rgba(255,255,255,0.8)", minWidth: 72 }}
              >
                {formatTime(currentTime)} / {formatTime(duration)}
              </Typography>
            )}

            <IconButton
              size={compact ? "small" : "medium"}
              aria-label={isMuted ? "开启声音" : "静音"}
              onClick={(event) => {
                event.stopPropagation();
                handleToggleMute();
              }}
              sx={{
                bgcolor: "rgba(255,255,255,0.12)",
                color: "common.white",
                "&:hover": {
                  bgcolor: "rgba(255,255,255,0.22)",
                },
              }}
            >
              {isMuted ? (
                <VolumeOffRounded fontSize={iconFontSize} />
              ) : (
                <VolumeUpRounded fontSize={iconFontSize} />
              )}
            </IconButton>
          </Box>
        </Box>
      )}
    </Box>
  );
};

export default VideoPlayer;
