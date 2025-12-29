import React, { useCallback, useEffect, useRef, useState } from "react";
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
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  IconButton,
} from "@mui/material";
import {
  Send,
  Download,
  AutoAwesome,
  Image as ImageIcon,
  OpenInFull,
} from "@mui/icons-material";
import { useLocation } from "react-router-dom";
import ReactPlayer from "react-player";

import type {
  GenerationRequest,
  GenerationResult,
  GenerationEventPayload,
  AIProvider,
  AIModel,
} from "../types";
import {
  generateContent,
  fetchProviders,
  fetchUsageRecordDetail,
  subscribeGenerationEvents,
} from "../ai";
import ImageUpload from "../components/ImageUpload";
import ImageViewer from "../components/ImageViewer";
import { buildDownloadName, isVideoUrl } from "../utils/media";

function downloadMedia(
  dataUrl: string,
  filename: string = "generated-media",
): void {
  try {
    const link = document.createElement("a");
    link.href = dataUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  } catch (error) {
    console.error("Error downloading image:", error);
    alert("下载失败，请重试");
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
  const [selectedProvider, setSelectedProvider] = useState<AIProvider | null>(
    null,
  );
  const [selectedModel, setSelectedModel] = useState<AIModel | null>(null);
  const [selectedSize, setSelectedSize] = useState<string>("");
  const [selectedDuration, setSelectedDuration] = useState<number | "">("");
  const [resultViewerOpen, setResultViewerOpen] = useState(false);
  const [resultViewerIndex, setResultViewerIndex] = useState(0);
  const [selectionInitialized, setSelectionInitialized] = useState(false);
  const clientIdRef = useRef<string>("");
  const activeRecordIdRef = useRef<number | null>(null);

  const ensureClientId = useCallback((): string => {
    if (clientIdRef.current) {
      return clientIdRef.current;
    }
    const generated =
      typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
        ? crypto.randomUUID()
        : `client-${Date.now()}-${Math.random().toString(16).slice(2)}`;
    clientIdRef.current = generated;
    return generated;
  }, []);

  const resolutionOptions = React.useMemo(() => {
    if (!selectedModel) {
      return [] as string[];
    }
    return selectedModel.supported_sizes ?? [];
  }, [selectedModel]);

  const durationOptions = React.useMemo(() => {
    if (!selectedModel) {
      return [] as number[];
    }
    return selectedModel.supported_durations ?? [];
  }, [selectedModel]);

  const isVideoModel = React.useMemo(() => {
    const id = selectedModel?.model_id?.toLowerCase() ?? "";
    const outputModalities =
      selectedModel?.output_modalities?.map((m) => m.toLowerCase()) ??
      selectedModel?.input_modalities?.map((m) => m.toLowerCase()) ??
      [];
    if (outputModalities.includes("video")) {
      return true;
    }
    if (
      selectedModel?.supported_durations?.length ||
      selectedModel?.default_duration
    ) {
      return true;
    }
    return id.includes("i2v") || id.includes("video");
  }, [selectedModel]);

  const handleStreamEvent = useCallback(
    async (event: GenerationEventPayload) => {
      try {
        const detail = await fetchUsageRecordDetail(event.record_id);
        const outputs = (detail.output_images ?? [])
          .map((item) => item.url || item.path)
          .filter((src): src is string => Boolean(src));
        onGenerate({
          outputs,
          text: detail.output_text || undefined,
          record_id: detail.id,
          error: detail.error_message || event.error,
        });
        if (event.status === "failure" || detail.error_message) {
          setError(detail.error_message || event.error || "生成失败，请重试");
        } else {
          setError(null);
        }
        setResultViewerOpen(false);
        if (outputs.length > 0) {
          setResultViewerIndex(0);
        }
      } catch (err) {
        const message =
          event.error ??
          (err instanceof Error ? err.message : "加载生成详情失败");
        setError(message);
      } finally {
        if (activeRecordIdRef.current === event.record_id) {
          activeRecordIdRef.current = null;
        }
        setIsGenerating(false);
      }
    },
    [onGenerate],
  );

  useEffect(() => {
    const clientId = ensureClientId();
    const subscription = subscribeGenerationEvents(
      clientId,
      handleStreamEvent,
      (err) => {
        console.error("事件流连接异常", err);
        setError((prev) => prev ?? err.message);
        setIsGenerating((prev) => (prev ? false : prev));
        activeRecordIdRef.current = null;
      },
    );
    return () => {
      subscription.close();
    };
  }, [ensureClientId, handleStreamEvent]);

  useEffect(() => {
    setSelectionInitialized(false);
  }, [selectionKey]);

  useEffect(() => {
    const sizes = resolutionOptions;
    if (sizes.length === 0) {
      setSelectedSize("");
      return;
    }
    const defaultSize = selectedModel?.default_size;
    setSelectedSize((current) =>
      sizes.includes(current)
        ? current
        : defaultSize && sizes.includes(defaultSize)
          ? defaultSize
          : sizes[0],
    );
  }, [resolutionOptions, selectedModel]);

  useEffect(() => {
    const durations = durationOptions;
    if (durations.length === 0) {
      setSelectedDuration("");
      return;
    }
    const defaultDuration = selectedModel?.default_duration;
    setSelectedDuration((current) =>
      typeof current === "number" && durations.includes(current)
        ? current
        : defaultDuration && durations.includes(defaultDuration)
          ? defaultDuration
          : durations[0],
    );
  }, [durationOptions, selectedModel]);

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
          setSelectedSize("");
        }
        setSelectionInitialized(false);
      } catch (err) {
        if (!cancelled) {
          const message =
            err instanceof Error ? err.message : "获取模型列表失败";
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
      (initialProviderId
        ? providers.find((item) => item.id === initialProviderId)
        : undefined) ??
      providers[0] ??
      null;

    setSelectedProvider(provider);

    let model: AIModel | null = provider?.models?.[0] ?? null;
    if (provider && initialModelId) {
      const maybeModel = provider.models.find(
        (item) => item.model_id === initialModelId,
      );
      if (maybeModel) {
        model = maybeModel;
      }
    }

    setSelectedModel(model);

    let size = "";
    const sizeOptions = model?.supported_sizes ?? [];
    if (sizeOptions.length > 0) {
      const defaultSize = model?.default_size;
      if (initialSize && sizeOptions.includes(initialSize)) {
        size = initialSize;
      } else if (defaultSize && sizeOptions.includes(defaultSize)) {
        size = defaultSize;
      } else {
        size = sizeOptions[0];
      }
    }

    let duration: number | "" = "";
    const durationOptionsInit = model?.supported_durations ?? [];
    if (durationOptionsInit.length > 0) {
      const defaultDuration = model?.default_duration;
      if (
        typeof defaultDuration === "number" &&
        defaultDuration > 0 &&
        durationOptionsInit.includes(defaultDuration)
      ) {
        duration = defaultDuration;
      } else {
        duration = durationOptionsInit[0];
      }
    }

    setSelectedSize(size);
    setSelectedDuration(duration);
    setSelectionInitialized(true);
  }, [
    providers,
    initialProviderId,
    initialModelId,
    initialSize,
    selectionInitialized,
  ]);

  const handleGenerate = async () => {
    if (!prompt.trim()) {
      setError("请输入提示词");
      return;
    }

    if (!selectedProvider || !selectedModel) {
      setError("请选择可用的模型");
      return;
    }

    setIsGenerating(true);
    setError(null);
    setResultViewerOpen(false);
    activeRecordIdRef.current = null;

    try {
      const output: GenerationRequest["output"] = {};
      if (resolutionOptions.length > 0 && selectedSize) {
        output.size = selectedSize;
      }
      if (
        typeof selectedDuration === "number" &&
        selectedDuration > 0
      ) {
        output.duration = selectedDuration;
      }

      const input_media =
        images.length > 0
          ? images.map((content) => ({ type: "image", content }))
          : undefined;
      const request: GenerationRequest = {
        prompt: prompt.trim(),
        provider: selectedProvider,
        model: selectedModel.model_id,
        ...(input_media ? { input_media } : {}),
        ...(Object.keys(output).length > 0 ? { output } : {}),
      };

      const job = await generateContent(request, ensureClientId());
      activeRecordIdRef.current = job.record_id;
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "生成失败，请重试";
      setError(errorMessage);
      console.error(`生成失败: ${errorMessage}`);
      console.error("Generation error:", err);
      setIsGenerating(false);
    }
  };

  const handleDownload = (index: number) => {
    const image = lastGeneratedImages[index];
    if (image) {
      const filename = buildDownloadName(
        image,
        `generated-media-${index + 1}`,
        isVideoUrl(image) ? ".mp4" : ".png",
      );
      downloadMedia(image, filename);
    }
  };

  const handleDownloadAll = () => {
    lastGeneratedImages.forEach((image, index) => {
      const filename = buildDownloadName(
        image,
        `generated-media-${index + 1}`,
        isVideoUrl(image) ? ".mp4" : ".png",
      );
      downloadMedia(image, filename);
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
        p: { xs: 2, sm: 2.5 },
        borderRadius: { xs: 2, sm: 2 },
        border: "1px solid",
        borderColor: "divider",
        boxShadow: {
          xs: "0 1px 3px rgba(15,23,42,0.04)",
          sm: "0 4px 6px -1px rgba(15,23,42,0.06)",
        },
        bgcolor: "background.paper",
      }}
    >
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          mb: { xs: 2, sm: 2 },
        }}
      >
        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <AutoAwesome color="primary" fontSize="small" />
          <Typography variant="h6" component="h2" sx={{ fontWeight: 600, fontSize: "1.1rem" }}>
            AI 图片生成器
          </Typography>
        </Box>
        {providersLoading && <CircularProgress size={18} />}
      </Box>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: {
            xs: "1fr",
            md: "minmax(112px, 168px) minmax(0,1fr)",
          },
          gap: { xs: 2, md: 2.5 },
          alignItems: "stretch",
          mb: { xs: 2, sm: 2 },
        }}
      >
        <ImageUpload
          images={images}
          onImagesChange={onImagesChange}
          maxImages={5}
          variant="inline"
        />

        <Box
          sx={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            gap: { xs: 1.5, sm: 1.5 },
          }}
        >
          <TextField
            fullWidth
            multiline
            minRows={3}
            maxRows={10}
            label="创意描述"
            placeholder="详细描述你想要的图片..."
            value={prompt}
            onChange={(e) => onPromptChange(e.target.value)}
            disabled={isGenerating}
            size="small"
            InputProps={{
              sx: {
                alignItems: "flex-start",
                px: { xs: 1.25, sm: 1.5 },
                py: { xs: 1, sm: 1.25 },
                "& .MuiInputBase-inputMultiline": {
                  fontSize: { xs: "0.875rem", sm: "0.9rem" },
                  lineHeight: 1.5,
                },
              },
            }}
            FormHelperTextProps={{
              sx: {
                display: "flex",
                justifyContent: "space-between",
                flexWrap: "wrap",
                gap: 1,
                mt: 0.5,
              },
            }}
            helperText={
              selectedProvider && selectedModel
                ? `${prompt.length}/500 字符 · ${selectedProvider.name} · ${selectedModel.name}`
                : `${prompt.length}/500 字符`
            }
          />
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: {
                xs: "repeat(auto-fit, minmax(130px, 1fr))",
                sm: "repeat(auto-fit, minmax(160px, 1fr))",
              },
              gap: { xs: 1, sm: 1.5 },
            }}
          >
            <FormControl fullWidth size="small">
              <InputLabel>AI 服务商</InputLabel>
              <Select
                value={selectedProvider?.id ?? ""}
                label="AI 服务商"
                disabled={
                  isGenerating || providersLoading || providers.length === 0
                }
                onChange={(e) => {
                  const provider = providers.find(
                    (p) => p.id === e.target.value,
                  );
                  if (!provider) {
                    return;
                  }
                  setSelectedProvider(provider);
                  const nextModel = provider.models[0] ?? null;
                  setSelectedModel(nextModel);
                  const nextSizes = nextModel?.supported_sizes ?? [];
                  const nextDefaultSize = nextModel?.default_size;
                  const fallbackSize =
                    nextSizes.length > 0 ? nextSizes[0] : "";
                  const resolvedSize =
                    nextDefaultSize && nextSizes.includes(nextDefaultSize)
                      ? nextDefaultSize
                      : fallbackSize;
                  setSelectedSize(resolvedSize);
                  const nextDurations =
                    nextModel?.supported_durations ?? [];
                  const nextDefaultDuration = nextModel?.default_duration;
                  const fallbackDuration =
                    nextDurations.length > 0 ? nextDurations[0] : "";
                  const resolvedDuration =
                    nextDefaultDuration &&
                      nextDurations.includes(nextDefaultDuration)
                      ? nextDefaultDuration
                      : fallbackDuration;
                  setSelectedDuration(resolvedDuration);
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
                value={selectedModel?.model_id ?? ""}
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
                  const model = selectedProvider.models.find(
                    (m) => m.model_id === e.target.value,
                  );
                  if (!model) {
                    return;
                  }
                  setSelectedModel(model);
                  const nextSizes = model.supported_sizes ?? [];
                  const nextDefaultSize = model.default_size;
                  const fallbackSize =
                    nextSizes.length > 0 ? nextSizes[0] : "";
                  const resolvedSize =
                    nextDefaultSize && nextSizes.includes(nextDefaultSize)
                      ? nextDefaultSize
                      : fallbackSize;
                  setSelectedSize(resolvedSize);
                  const nextDurations = model.supported_durations ?? [];
                  const nextDefaultDuration = model.default_duration;
                  const fallbackDuration =
                    nextDurations.length > 0 ? nextDurations[0] : "";
                  const resolvedDuration =
                    nextDefaultDuration &&
                      nextDurations.includes(nextDefaultDuration)
                      ? nextDefaultDuration
                      : fallbackDuration;
                  setSelectedDuration(resolvedDuration);
                }}
              >
                {(selectedProvider?.models ?? []).map((model) => (
                  <MenuItem key={model.id} value={model.model_id}>
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

            {resolutionOptions.length > 0 && (
              <FormControl fullWidth size="small">
                <InputLabel>{isVideoModel ? "分辨率" : "图片尺寸"}</InputLabel>
                <Select
                  value={selectedSize}
                  label={isVideoModel ? "分辨率" : "图片尺寸"}
                  disabled={isGenerating}
                  onChange={(e) => setSelectedSize(e.target.value)}
                >
                  {resolutionOptions.map((sizeOption) => (
                    <MenuItem key={sizeOption} value={sizeOption}>
                      {sizeOption}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            )}

            {durationOptions.length > 0 && (
              <FormControl fullWidth size="small">
                <InputLabel>视频时长</InputLabel>
                <Select
                  value={selectedDuration === "" ? "" : selectedDuration}
                  label="视频时长"
                  disabled={isGenerating}
                  onChange={(e) => {
                    const value = Number(e.target.value);
                    setSelectedDuration(Number.isFinite(value) ? value : "");
                  }}
                >
                  {durationOptions.map((option) => (
                    <MenuItem key={option} value={option}>
                      {option} 秒
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
          <Box
            component="pre"
            sx={{ whiteSpace: "pre-wrap", fontFamily: "inherit", m: 0 }}
          >
            {error}
          </Box>
        </Alert>
      )}

      <Box
        sx={{
          display: "flex",
          flexDirection: { xs: "column", sm: "row" },
          justifyContent: { xs: "stretch", sm: "flex-end" },
          alignItems: { xs: "stretch", sm: "center" },
          gap: 1,
          mb: { xs: 2, sm: 2 },
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
          startIcon={
            isGenerating ? (
              <CircularProgress size={16} color="inherit" />
            ) : (
              <Send fontSize="small" />
            )
          }
          sx={{
            width: { xs: "100%", sm: "auto" },
            minWidth: { sm: 160 },
            py: 1,
            fontWeight: 600,
            borderRadius: 999,
            background: "linear-gradient(45deg, #FF6B6B 30%, #4ECDC4 90%)",
            "&:hover": {
              background: "linear-gradient(45deg, #FF5252 30%, #26A69A 90%)",
            },
          }}
        >
          {isGenerating ? "创作中..." : "开始生成"}
        </Button>
      </Box>

      {lastGeneratedImages.length > 0 && (
        <Card
          sx={{
            mb: { xs: 2, sm: 2 },
            borderRadius: { xs: 2, sm: 2 },
            boxShadow: "0 4px 12px rgba(15,23,42,0.08)",
          }}
        >
          <Box
            sx={{
              p: { xs: 1.5, sm: 2 },
              display: "flex",
              flexDirection: { xs: "column", sm: "row" },
              gap: { xs: 1, sm: 2 },
              alignItems: { xs: "flex-start", sm: "center" },
              justifyContent: "space-between",
            }}
          >
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <ImageIcon color="success" fontSize="small" />
              <Typography variant="body2" fontWeight="medium">
                生成成功
              </Typography>
              <Chip
                label={`共 ${lastGeneratedImages.length} 张`}
                color="success"
                size="small"
              />
            </Box>
            <Button
              variant="contained"
              color="success"
              size="small"
              startIcon={<Download />}
              onClick={
                lastGeneratedImages.length > 1
                  ? handleDownloadAll
                  : () => handleDownload(0)
              }
              sx={{ borderRadius: 999, px: 2.5 }}
            >
              {lastGeneratedImages.length > 1 ? "下载全部" : "下载"}
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
              display: "grid",
              gridTemplateColumns: {
                xs: "repeat(auto-fit, minmax(160px, 1fr))",
                sm: "repeat(auto-fit, minmax(220px, 1fr))",
              },
              gap: { xs: 1.5, sm: 2 },
              px: { xs: 1.5, sm: 2 },
              pb: { xs: 1.5, sm: 2 },
            }}
          >
            {lastGeneratedImages.map((image, index) => {
              const isVideo = isVideoUrl(image);
              return (
                <Box
                  key={image + index}
                  sx={{
                    display: "flex",
                    flexDirection: "column",
                    gap: 1,
                  }}
                >
                  <Box
                    sx={{
                      position: "relative",
                      width: "100%",
                      aspectRatio: { xs: "4 / 5", sm: "16 / 9" },
                      maxHeight: 420,
                      bgcolor: "grey.100",
                      borderRadius: 2,
                      overflow: "hidden",
                      cursor: isVideo ? "default" : "pointer",
                      transition: "all 0.2s",
                      "& img, & .react-player": {
                        transition: "transform 0.2s ease",
                      },
                      "&:hover img, &:hover .react-player": {
                        transform: "scale(1.01)",
                      },
                    }}
                    onClick={
                      isVideo ? undefined : () => handleOpenResultViewer(index)
                    }
                  >
                    {isVideo ? (
                      <>
                        <ReactPlayer
                          src={image}
                          controls
                          width="100%"
                          height="100%"
                          className="react-player"
                          style={{
                            position: "absolute",
                            inset: 0,
                          }}
                          playsInline
                          config={{ html: { controlsList: "nodownload" } }}
                        />
                        <IconButton
                          size="small"
                          color="inherit"
                          onClick={() => handleOpenResultViewer(index)}
                          sx={{
                            position: "absolute",
                            top: 8,
                            right: 8,
                            bgcolor: "rgba(0,0,0,0.55)",
                            color: "common.white",
                            "&:hover": {
                              bgcolor: "rgba(0,0,0,0.7)",
                            },
                          }}
                          aria-label={`放大预览第 ${index + 1} 个视频`}
                        >
                          <OpenInFull fontSize="small" />
                        </IconButton>
                      </>
                    ) : (
                      <Box
                        component="img"
                        src={image}
                        alt={`AI生成的媒体 ${index + 1}`}
                        sx={{
                          width: "100%",
                          height: "100%",
                          objectFit: "cover",
                          display: "block",
                        }}
                      />
                    )}
                  </Box>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<Download />}
                    onClick={() => handleDownload(index)}
                    sx={{ alignSelf: "flex-end" }}
                  >
                    下载第 {index + 1} {isVideo ? "个视频" : "张图片"}
                  </Button>
                </Box>
              );
            })}
          </Box>
        </Card>
      )}

      <ImageViewer
        open={resultViewerOpen}
        onClose={handleCloseViewer}
        imageUrl={lastGeneratedImages[resultViewerIndex] || ""}
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
  const [prompt, setPrompt] = useState("");
  const [images, setImages] = useState<string[]>([]);
  const [lastGeneratedImages, setLastGeneratedImages] = useState<string[]>([]);
  const [lastGenerationText, setLastGenerationText] = useState<string | null>(
    null,
  );
  const [initialSelection, setInitialSelection] = useState<{
    providerId?: string;
    modelId?: string;
    size?: string;
  }>({});
  const [selectionKey, setSelectionKey] = useState<string>("");
  const lastAppliedPrefillKeyRef = useRef<string | null>(null);

  useEffect(() => {
    const state = location.state as HistoryPrefillState | null | undefined;

    if (!state) {
      return;
    }

    if (lastAppliedPrefillKeyRef.current === location.key) {
      return;
    }

    setPrompt(state.prompt ?? "");
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

  const handleGenerate = useCallback((result: GenerationResult) => {
    setLastGeneratedImages(result.outputs);
    setLastGenerationText(result.text ?? result.error ?? null);
  }, []);

  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100%" }}>
      <Container maxWidth="md" sx={{ py: { xs: 2, sm: 3 }, flex: 1 }}>
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            gap: { xs: 2, sm: 3 },
          }}
        >
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
