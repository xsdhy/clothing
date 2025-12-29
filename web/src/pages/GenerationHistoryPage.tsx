import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Container,
  FormControl,
  IconButton,
  InputLabel,
  MenuItem,
  Pagination,
  Select,
  Autocomplete,
  TextField,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";

import ReplayIcon from "@mui/icons-material/Replay";
import AddPhotoAlternateIcon from "@mui/icons-material/AddPhotoAlternate";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import { useNavigate } from "react-router-dom";
import VideoPlayer from "../components/VideoPlayer";

import {
  fetchUsageRecords,
  fetchProviders,
  fetchTags,
  updateUsageRecordTags,
  deleteUsageRecord,
  type UsageRecordsResult,
  type UsageRecordResultFilter,
} from "../ai";
import type { AIProvider, UsageRecord, Tag } from "../types";
import ImageViewer from "../components/ImageViewer";
import UsageRecordDetailDialog from "../components/UsageRecordDetailDialog";
import { useAuth } from "../contexts/AuthContext";
import { isVideoUrl } from "../utils/media";

const PAGE_SIZE = 10;
const ALL_VALUE = "all";
const DEFAULT_RESULT_FILTER: UsageRecordResultFilter = "success";

type ModelOption = { id: string; name: string };

const GenerationHistoryPage: React.FC = () => {
  const { isAdmin, user } = useAuth();
  const [records, setRecords] = useState<UsageRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filtersError, setFiltersError] = useState<string | null>(null);
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [providersLoading, setProvidersLoading] = useState(false);
  const [providerFilter, setProviderFilter] = useState<string>(ALL_VALUE);
  const [modelFilter, setModelFilter] = useState<string>(ALL_VALUE);
  const [resultFilter, setResultFilter] = useState<UsageRecordResultFilter>(
    DEFAULT_RESULT_FILTER,
  );
  const [tags, setTags] = useState<Tag[]>([]);
  const [tagsLoading, setTagsLoading] = useState(false);
  const [tagFilter, setTagFilter] = useState<number[]>([]);
  const [tagsLoadError, setTagsLoadError] = useState<string | null>(null);
  const [tagSavingRecordId, setTagSavingRecordId] = useState<number | null>(
    null,
  );
  const [tagActionError, setTagActionError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState<UsageRecordsResult["meta"] | null>(null);
  const [previewImage, setPreviewImage] = useState<{
    url: string;
    alt: string;
  } | null>(null);
  const [retryingRecordId, setRetryingRecordId] = useState<number | null>(null);
  const [preparingOutputAsInput, setPreparingOutputAsInput] = useState<{
    recordId: number;
    index: number;
  } | null>(null);
  const [selectedDetail, setSelectedDetail] = useState<{
    recordId: number;
    imageIndex?: number;
  } | null>(null);
  const [deletingRecordId, setDeletingRecordId] = useState<number | null>(null);
  const navigate = useNavigate();
  const recordCount = records.length;

  const totalPages = useMemo(() => {
    if (!meta || meta.page_size <= 0) {
      return 1;
    }
    return Math.max(1, Math.ceil(meta.total / meta.page_size));
  }, [meta]);

  const availableModels = useMemo<ModelOption[]>(() => {
    if (providerFilter === ALL_VALUE) {
      const result = new Map<string, string>();
      providers.forEach((provider) => {
        provider.models.forEach((model) => {
          if (!result.has(model.model_id)) {
            result.set(model.model_id, model.name ?? model.model_id);
          }
        });
      });
      return Array.from(result.entries()).map(([id, name]) => ({ id, name }));
    }

    const provider = providers.find((item) => item.id === providerFilter);
    if (!provider) {
      return [];
    }

    return provider.models.map((model) => ({
      id: model.model_id,
      name: model.name ?? model.model_id,
    }));
  }, [providerFilter, providers]);

  const selectedTagOptions = useMemo(
    () => tags.filter((tag) => tagFilter.includes(tag.id)),
    [tagFilter, tags],
  );

  const loadRecords = useCallback(
    async (targetPage: number) => {
      setLoading(true);
      setError(null);
      try {
        const result = await fetchUsageRecords(targetPage, PAGE_SIZE, {
          provider: providerFilter !== ALL_VALUE ? providerFilter : undefined,
          model: modelFilter !== ALL_VALUE ? modelFilter : undefined,
          result: resultFilter,
          tags: tagFilter,
        });
        setRecords(result.records);
        setMeta(result.meta);

        if (result.meta) {
          if (result.meta.total === 0 && targetPage !== 1) {
            setPage(1);
          } else if (result.meta.page > 0 && result.meta.page !== targetPage) {
            setPage(result.meta.page);
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "获取记录失败");
      } finally {
        setLoading(false);
      }
    },
    [modelFilter, providerFilter, resultFilter, tagFilter],
  );

  useEffect(() => {
    void loadRecords(page);
  }, [loadRecords, page]);

  useEffect(() => {
    if (!selectedDetail) {
      return;
    }
    if (!records.some((record) => record.id === selectedDetail.recordId)) {
      setSelectedDetail(null);
    }
  }, [records, selectedDetail]);

  useEffect(() => {
    let cancelled = false;

    const loadProviders = async () => {
      setProvidersLoading(true);
      setFiltersError(null);
      try {
        const list = await fetchProviders();
        if (!cancelled) {
          setProviders(list);
        }
      } catch (err) {
        if (!cancelled) {
          setFiltersError(
            err instanceof Error ? err.message : "获取模型提供商失败",
          );
        }
      } finally {
        if (!cancelled) {
          setProvidersLoading(false);
        }
      }
    };

    void loadProviders();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;

    const loadTags = async () => {
      setTagsLoading(true);
      setTagsLoadError(null);
      try {
        const list = await fetchTags();
        if (!cancelled) {
          setTags(list);
        }
      } catch (err) {
        if (!cancelled) {
          setTagsLoadError(err instanceof Error ? err.message : "获取标签失败");
        }
      } finally {
        if (!cancelled) {
          setTagsLoading(false);
        }
      }
    };

    void loadTags();

    return () => {
      cancelled = true;
    };
  }, []);

  const handleRefresh = useCallback(() => {
    void loadRecords(page);
  }, [loadRecords, page]);

  const handleProviderChange = useCallback(
    (event: SelectChangeEvent<string>) => {
      const value = event.target.value || ALL_VALUE;
      setProviderFilter(value);
      setModelFilter(ALL_VALUE);
      setPage(1);
    },
    [],
  );

  const handleModelChange = useCallback((event: SelectChangeEvent<string>) => {
    const value = event.target.value || ALL_VALUE;
    setModelFilter(value);
    setPage(1);
  }, []);

  const handleResultChange = useCallback((event: SelectChangeEvent<string>) => {
    const value = event.target.value as UsageRecordResultFilter;
    const nextValue: UsageRecordResultFilter =
      value === "failure" || value === "all" ? value : DEFAULT_RESULT_FILTER;
    setResultFilter(nextValue);
    setPage(1);
  }, []);

  const handleTagFilterChange = useCallback((_event: unknown, value: Tag[]) => {
    const ids = value
      .map((tag) => tag.id)
      .filter((id) => Number.isFinite(id) && id > 0);
    setTagFilter(ids);
    setPage(1);
  }, []);

  const handleDelete = useCallback(
    async (record: UsageRecord) => {
      if (!window.confirm("确认删除该生成记录吗？")) {
        return;
      }

      try {
        setDeletingRecordId(record.id);
        const nextPage = recordCount === 1 && page > 1 ? page - 1 : page;
        await deleteUsageRecord(record.id);
        setSelectedDetail((prev) =>
          prev?.recordId === record.id ? null : prev,
        );

        if (nextPage !== page) {
          setPage(nextPage);
        } else {
          await loadRecords(page);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "删除生成记录失败");
      } finally {
        setDeletingRecordId(null);
      }
    },
    [loadRecords, page, recordCount],
  );

  const handlePageChange = useCallback(
    (_event: React.ChangeEvent<unknown>, value: number) => {
      setPage(value);
    },
    [],
  );

  const handleImageClick = useCallback((url: string, alt: string) => {
    setPreviewImage({ url, alt });
  }, []);

  const handleClosePreview = useCallback(() => {
    setPreviewImage(null);
  }, []);

  const handleDownloadPreview = useCallback(
    (_index: number, item: { src: string }) => {
      const url = item?.src ?? previewImage?.url;
      if (!url) {
        return;
      }

      const link = document.createElement("a");
      link.href = url;
      const fileName = (previewImage?.alt ?? "image").replace(/\s+/g, "_");
      link.download = fileName;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
    [previewImage],
  );

  const loadImageAsDataUrl = useCallback(
    async (url: string): Promise<string> => {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(
          `无法读取图片: ${response.status} ${response.statusText}`,
        );
      }

      const blob = await response.blob();

      return await new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
          const result = reader.result;
          if (typeof result === "string") {
            resolve(result);
          } else {
            reject(new Error("图片转换失败"));
          }
        };
        reader.onerror = () => {
          reject(new Error("图片读取失败"));
        };
        reader.readAsDataURL(blob);
      });
    },
    [],
  );

  const handleRetry = useCallback(
    async (record: UsageRecord) => {
      try {
        setRetryingRecordId(record.id);
        const base64Images = await Promise.all(
          record.input_images.map((image) => loadImageAsDataUrl(image.url)),
        );

        const state = {
          prompt: record.prompt,
          inputImages: base64Images,
          providerId: record.provider_id,
          modelId: record.model_id,
          size: record.size,
        } as const;

        navigate("/custom", { state });
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "重新创建生成任务失败";
        setError(message);
      } finally {
        setRetryingRecordId(null);
      }
    },
    [loadImageAsDataUrl, navigate],
  );

  const handleUseOutputImage = useCallback(
    async (record: UsageRecord, imageUrl: string, index: number) => {
      try {
        setPreparingOutputAsInput({ recordId: record.id, index });
        const base64Image = await loadImageAsDataUrl(imageUrl);

        const state = {
          prompt: record.prompt,
          inputImages: [base64Image],
          providerId: record.provider_id,
          modelId: record.model_id,
          size: record.size,
        } as const;

        navigate("/custom", { state });
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "无法将图片带入生成页面";
        setError(message);
      } finally {
        setPreparingOutputAsInput(null);
      }
    },
    [loadImageAsDataUrl, navigate],
  );

  const handleOpenDetails = useCallback((record: UsageRecord) => {
    setSelectedDetail({
      recordId: record.id,
      imageIndex: record.output_images.length > 0 ? 0 : undefined,
    });
  }, []);

  const handleCloseDetail = useCallback(() => {
    setSelectedDetail(null);
  }, []);

  const handleTagsUpdated = useCallback((updated: UsageRecord) => {
    setRecords((prev) =>
      prev.map((item) => (item.id === updated.id ? updated : item)),
    );
  }, []);

  const mergeTagsOptions = useCallback(
    (items: Tag[]) => {
      setTags((prev) => {
        const map = new Map<number, Tag>();
        prev.forEach((tag) => map.set(tag.id, tag));
        items.forEach((tag) => map.set(tag.id, tag));
        return Array.from(map.values()).sort((a, b) =>
          a.name.localeCompare(b.name),
        );
      });
    },
    [setTags],
  );

  const handleUpdateRecordTags = useCallback(
    async (record: UsageRecord, nextTags: Tag[]) => {
      const canEdit =
        isAdmin || (user && record.user && record.user.id === user.id);
      if (!canEdit) {
        return;
      }

      setTagSavingRecordId(record.id);
      setTagActionError(null);
      const ids = nextTags
        .map((tag) => tag.id)
        .filter((id) => Number.isFinite(id) && id > 0);
      try {
        const updated = await updateUsageRecordTags(record.id, ids);
        mergeTagsOptions(updated.tags ?? []);
        setRecords((prev) =>
          prev.map((item) => (item.id === updated.id ? updated : item)),
        );
      } catch (err) {
        setTagActionError(
          err instanceof Error ? err.message : "更新标签失败",
        );
      } finally {
        setTagSavingRecordId(null);
      }
    },
    [isAdmin, mergeTagsOptions, user],
  );

  const selectedRecordId = selectedDetail?.recordId ?? null;
  const preparingMatch =
    selectedRecordId !== null &&
    preparingOutputAsInput?.recordId === selectedRecordId;
  const preparingImageIndex = preparingMatch
    ? preparingOutputAsInput?.index
    : undefined;

  return (
    <Box sx={{ py: 3 }}>
      <Container maxWidth="lg">
        <Stack
          direction={{ xs: "column", sm: "row" }}
          alignItems={{ xs: "flex-start", sm: "center" }}
          justifyContent="space-between"
          spacing={2}
          sx={{ mb: 2.5 }}
        >
          <Box>
            <Typography variant="h5" gutterBottom sx={{ mb: 0.5, fontWeight: 700 }}>
              生成记录
            </Typography>
            <Typography variant="body2" color="text.secondary">
              查看每次调用模型的输入提示、生成结果与可能的错误信息。
            </Typography>
          </Box>
          <Button
            variant="contained"
            color="primary"
            onClick={handleRefresh}
            disabled={loading}
          >
            刷新
          </Button>
        </Stack>

        <Card elevation={0} sx={{ mb: 3 }}>
          <CardContent>
            <Stack
              direction={{ xs: "column", md: "row" }}
              spacing={2}
              alignItems={{ xs: "stretch", md: "center" }}
            >
              <FormControl
                size="small"
                sx={{ minWidth: 160 }}
                disabled={providersLoading && providers.length === 0}
              >
                <InputLabel id="filter-provider-label">厂商</InputLabel>
                <Select
                  labelId="filter-provider-label"
                  value={providerFilter}
                  label="厂商"
                  onChange={handleProviderChange}
                >
                  <MenuItem value={ALL_VALUE}>全部厂商</MenuItem>
                  {providers.map((provider) => (
                    <MenuItem key={provider.id} value={provider.id}>
                      {provider.name ?? provider.id}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl
                size="small"
                sx={{ minWidth: 160 }}
                disabled={
                  providersLoading &&
                  providers.length === 0 &&
                  availableModels.length === 0
                }
              >
                <InputLabel id="filter-model-label">模型</InputLabel>
                <Select
                  labelId="filter-model-label"
                  value={modelFilter}
                  label="模型"
                  onChange={handleModelChange}
                >
                  <MenuItem value={ALL_VALUE}>全部模型</MenuItem>
                  {availableModels.map((model) => (
                    <MenuItem key={model.id} value={model.id}>
                      {model.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl size="small" sx={{ minWidth: 160 }}>
                <InputLabel id="filter-result-label">生成结果</InputLabel>
                <Select
                  labelId="filter-result-label"
                  value={resultFilter}
                  label="生成结果"
                  onChange={handleResultChange}
                >
                  <MenuItem value="success">成功</MenuItem>
                  <MenuItem value="failure">失败</MenuItem>
                  <MenuItem value="all">全部</MenuItem>
                </Select>
              </FormControl>

              <Autocomplete
                multiple
                size="small"
                options={tags}
                value={selectedTagOptions}
                onChange={handleTagFilterChange}
                loading={tagsLoading}
                sx={{ minWidth: 220, flex: 1 }}
                getOptionLabel={(option) => option.name}
                isOptionEqualToValue={(option, value) => option.id === value.id}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label="标签"
                    placeholder="按标签筛选"
                  />
                )}
              />

              {providersLoading && (
                <Stack
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  sx={{ minHeight: 40 }}
                >
                  <CircularProgress size={18} />
                  <Typography variant="caption" color="text.secondary">
                    正在加载可用厂商
                  </Typography>
                </Stack>
              )}
            </Stack>

            {filtersError && (
              <Alert severity="warning" sx={{ mt: 2 }}>
                {filtersError}
              </Alert>
            )}
            {tagsLoadError && (
              <Alert severity="warning" sx={{ mt: 2 }}>
                {tagsLoadError}
              </Alert>
            )}
          </CardContent>
        </Card>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}
        {tagActionError && (
          <Alert severity="warning" sx={{ mb: 3 }}>
            {tagActionError}
          </Alert>
        )}

        {loading && (
          <Stack
            direction="row"
            spacing={2}
            alignItems="center"
            justifyContent="center"
            sx={{ my: 6 }}
          >
            <CircularProgress size={28} />
            <Typography variant="body2" color="text.secondary">
              正在加载生成记录…
            </Typography>
          </Stack>
        )}

        {!loading && records.length === 0 && !error && (
          <Box sx={{ textAlign: "center", py: 8 }}>
            <Typography variant="h6" gutterBottom>
              暂无生成记录
            </Typography>
            <Typography variant="body2" color="text.secondary">
              完成一次图片生成后，可在此处查看完整的调用详情。
            </Typography>
          </Box>
        )}

        <Stack spacing={2} sx={{ opacity: loading ? 0.7 : 1 }}>
          {records.map((record) => {
            const hasError = Boolean(record.error_message);
            const canEditRecordTags =
              isAdmin || (user && record.user && record.user.id === user.id);
            const recordTagOptionsMap = new Map<number, Tag>();
            tags.forEach((tag) => recordTagOptionsMap.set(tag.id, tag));
            (record.tags ?? []).forEach((tag) =>
              recordTagOptionsMap.set(tag.id, tag),
            );
            const recordTagOptions = Array.from(recordTagOptionsMap.values());
            const selectedRecordTags = recordTagOptions.filter((tag) =>
              (record.tags ?? []).some((t) => t.id === tag.id),
            );
            const imageList = [
              ...record.input_images.map((img, idx) => ({
                ...img,
                alt: `I${idx + 1}`,
                isOutput: false,
                index: idx,
              })),
              ...record.output_images.map((img, idx) => ({
                ...img,
                alt: `O${idx + 1}`,
                isOutput: true,
                index: idx,
              })),
            ];

            return (
              <Card
                key={record.id}
                elevation={1}
                sx={{
                  overflow: "hidden",
                }}
              >
                <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
                  <Stack
                    direction={{ xs: "column", md: "row" }}
                    spacing={2}
                    alignItems={{ xs: "stretch", md: "flex-start" }}
                  >
                    <Box
                      sx={{
                        display: "grid",
                        gridTemplateColumns: "repeat(auto-fill, minmax(120px, 1fr))",
                        gap: 1,
                        flex: { md: "0 0 380px" },
                        maxWidth: { md: 480 },
                      }}
                    >
                      {imageList.length === 0 && (
                        <Box
                          sx={{
                            border: "1px dashed",
                            borderColor: "divider",
                            borderRadius: 2,
                            p: 2,
                            textAlign: "center",
                            color: "text.secondary",
                          }}
                        >
                          无图片
                        </Box>
                      )}
                      {imageList.map((image) => {
                        const isVideo = isVideoUrl(image.url);
                        return (
                          <Box
                            key={`${record.id}-${image.isOutput ? "out" : "in"}-${image.index}`}
                            sx={{
                              width: "100%",
                              borderRadius: 1.5,
                              overflow: "hidden",
                              border: "1px solid",
                              borderColor: "divider",
                              bgcolor: "background.paper",
                              position: "relative",
                              aspectRatio: "1/1",
                            }}
                          >
                            {isVideo ? (
                              <VideoPlayer
                                src={image.url}
                                compact
                                onOpen={() => handleImageClick(image.url, image.alt)}
                                openLabel={`预览 ${image.alt} 视频`}
                                sx={{ width: "100%", height: "100%", borderRadius: 0 }}
                              />
                            ) : (
                              <Box
                                component="img"
                                src={image.url}
                                alt={image.alt}
                                loading="lazy"
                                onClick={() => handleImageClick(image.url, image.alt)}
                                sx={{
                                  width: "100%",
                                  height: "100%",
                                  objectFit: "cover",
                                  cursor: "pointer",
                                }}
                              />
                            )}
                            <Chip
                              label={image.alt}
                              size="small"
                              color={image.isOutput ? "primary" : "default"}
                              sx={{
                                position: "absolute",
                                top: 6,
                                left: 6,
                                bgcolor: image.isOutput
                                  ? "primary.main"
                                  : "rgba(0,0,0,0.6)",
                                color: "common.white",
                              }}
                            />
                            {image.isOutput && (
                              <Tooltip title="作为输入图片使用">
                                <span>
                                  <IconButton
                                    size="small"
                                    color="primary"
                                    sx={{
                                      position: "absolute",
                                      top: 6,
                                      right: 6,
                                      bgcolor: "rgba(255,255,255,0.9)",
                                      "&:hover": {
                                        bgcolor: "rgba(255,255,255,1)",
                                      },
                                    }}
                                    onClick={(event) => {
                                      event.stopPropagation();
                                      void handleUseOutputImage(
                                        record,
                                        image.url,
                                        image.index,
                                      );
                                    }}
                                    disabled={
                                      preparingOutputAsInput?.recordId ===
                                      record.id &&
                                      preparingOutputAsInput?.index ===
                                      image.index
                                    }
                                  >
                                    {preparingOutputAsInput?.recordId ===
                                      record.id &&
                                      preparingOutputAsInput?.index ===
                                      image.index ? (
                                      <CircularProgress size={16} color="inherit" />
                                    ) : (
                                      <AddPhotoAlternateIcon fontSize="small" />
                                    )}
                                  </IconButton>
                                </span>
                              </Tooltip>
                            )}
                          </Box>
                        );
                      })}
                    </Box>

                    <Stack spacing={1.5} sx={{ flex: 1 }}>
                      <Stack
                        direction="row"
                        spacing={1}
                        alignItems="center"
                        flexWrap="wrap"
                        useFlexGap
                      >
                        <Chip
                          color={hasError ? "error" : "success"}
                          variant="outlined"
                          label={hasError ? "生成异常" : "生成成功"}
                        />
                        <Typography
                          variant="subtitle2"
                          color="text.secondary"
                        >
                          {new Date(record.created_at).toLocaleString()}
                        </Typography>
                        <Typography
                          variant="subtitle2"
                          color="text.secondary"
                        >
                          {record.provider_id}/{record.model_id}
                        </Typography>
                      </Stack>

                      <Typography variant="body2" color="text.primary">
                        提示词：{record.prompt || "（无）"}
                      </Typography>
                      {record.output_text && (
                        <Typography variant="body2" color="text.secondary">
                          文本输出：{record.output_text}
                        </Typography>
                      )}
                      {record.error_message && (
                        <Alert severity="error" variant="outlined">
                          {record.error_message}
                        </Alert>
                      )}

                      <Stack spacing={1}>


                        {canEditRecordTags && (
                          <Autocomplete
                            multiple
                            size="small"
                            options={recordTagOptions}
                            value={selectedRecordTags}
                            onChange={(_event, value) => {
                              void handleUpdateRecordTags(record, value);
                            }}
                            loading={tagsLoading}
                            disableCloseOnSelect
                            getOptionLabel={(option) => option.name ?? ""}
                            isOptionEqualToValue={(option, value) =>
                              option.id === value.id
                            }
                            renderTags={(value, getTagProps) =>
                              value.map((option, index) => (
                                <Chip
                                  {...getTagProps({ index })}
                                  key={option.id}
                                  size="small"
                                  label={option.name}
                                />
                              ))
                            }
                            renderInput={(params) => (
                              <TextField
                                {...params}
                                label="打标签"
                                placeholder="选择或搜索标签"
                              />
                            )}
                            disabled={
                              tagSavingRecordId === record.id || tagsLoading
                            }
                          />
                        )}
                      </Stack>

                      <Stack direction="row" spacing={1} flexWrap="wrap">
                        <Button
                          variant="contained"
                          size="small"
                          onClick={() => handleOpenDetails(record)}
                        >
                          查看详情
                        </Button>
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={
                            retryingRecordId === record.id ? (
                              <CircularProgress size={16} color="inherit" />
                            ) : (
                              <ReplayIcon fontSize="small" />
                            )
                          }
                          onClick={() => {
                            void handleRetry(record);
                          }}
                          disabled={retryingRecordId === record.id}
                        >
                          再来一次
                        </Button>
                        <Tooltip title="删除记录">
                          <span>
                            <Button
                              variant="text"
                              color="error"
                              size="small"
                              startIcon={
                                deletingRecordId === record.id ? (
                                  <CircularProgress size={16} color="inherit" />
                                ) : (
                                  <DeleteOutlineIcon fontSize="small" />
                                )
                              }
                              onClick={() => {
                                void handleDelete(record);
                              }}
                              disabled={deletingRecordId === record.id}
                            >
                              删除
                            </Button>
                          </span>
                        </Tooltip>
                      </Stack>
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>
            );
          })}
        </Stack>

        {totalPages > 1 && (
          <Stack alignItems="center" sx={{ mt: 4 }}>
            <Pagination
              count={totalPages}
              page={Math.min(page, totalPages)}
              onChange={handlePageChange}
              color="primary"
            />
            {meta && (
              <Typography
                variant="caption"
                color="text.secondary"
                sx={{ mt: 1 }}
              >
                共 {meta.total} 条记录，每页 {meta.page_size} 条
              </Typography>
            )}
          </Stack>
        )}

        <UsageRecordDetailDialog
          open={Boolean(selectedDetail)}
          recordId={selectedDetail?.recordId ?? null}
          initialImageIndex={selectedDetail?.imageIndex}
          onClose={handleCloseDetail}
          onRetry={(record) => {
            void handleRetry(record);
          }}
          onDelete={(record) => {
            void handleDelete(record);
          }}
          onUseOutputImage={(record, imageUrl, imageIndex) => {
            void handleUseOutputImage(record, imageUrl, imageIndex);
          }}
          onPreviewOutputImage={(record, imageIndex) => {
            const image = record.output_images[imageIndex];
            if (image?.url) {
              setPreviewImage({
                url: image.url,
                alt: `记录 #${record.id} 输出图片 ${imageIndex + 1}`,
              });
            }
          }}
          actionState={{
            retrying:
              selectedRecordId !== null &&
              retryingRecordId === selectedRecordId,
            deleting:
              selectedRecordId !== null &&
              deletingRecordId === selectedRecordId,
            preparingOutput: typeof preparingImageIndex === "number",
            preparingOutputIndex:
              typeof preparingImageIndex === "number"
                ? preparingImageIndex
                : undefined,
          }}
          onTagsUpdated={handleTagsUpdated}
        />

        <ImageViewer
          open={Boolean(previewImage)}
          onClose={handleClosePreview}
          imageUrl={previewImage?.url}
          title={previewImage?.alt}
          showDownload
          onDownload={handleDownloadPreview}
        />
      </Container>
    </Box>
  );
};

export default GenerationHistoryPage;
