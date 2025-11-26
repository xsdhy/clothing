import React, { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  ButtonBase,
  Chip,
  CircularProgress,
  Autocomplete,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import InputIcon from "@mui/icons-material/Input";
import ReplayIcon from "@mui/icons-material/Replay";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";


import { fetchUsageRecordDetail, fetchTags, updateUsageRecordTags } from "../ai";
import type { UsageRecord, Tag } from "../types";
import { isVideoUrl } from "../utils/media";
import { useAuth } from "../contexts/AuthContext";

const formatDateTime = (value: string): string => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};

export interface UsageRecordDetailDialogProps {
  open: boolean;
  recordId: number | null;
  initialImageIndex?: number;
  onClose: () => void;
  onRetry?: (record: UsageRecord) => void | Promise<void>;
  onDelete?: (record: UsageRecord) => void | Promise<void>;
  onUseOutputImage?: (
    record: UsageRecord,
    imageUrl: string,
    index: number,
  ) => void | Promise<void>;
  onPreviewOutputImage?: (record: UsageRecord, index: number) => void;
  onTagsUpdated?: (record: UsageRecord) => void;
  actionState?: {
    retrying?: boolean;
    deleting?: boolean;
    preparingOutput?: boolean;
    preparingOutputIndex?: number;
  };
}

const UsageRecordDetailDialog: React.FC<UsageRecordDetailDialogProps> = ({
  open,
  recordId,
  initialImageIndex,
  onClose,
  onRetry,
  onDelete,
  onUseOutputImage,
  onPreviewOutputImage,
  onTagsUpdated,
  actionState,
}) => {
  const { user, isAdmin } = useAuth();
  const [record, setRecord] = useState<UsageRecord | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedImageIndex, setSelectedImageIndex] = useState<
    number | undefined
  >(initialImageIndex);
  const [tagOptions, setTagOptions] = useState<Tag[]>([]);
  const [tagsLoading, setTagsLoading] = useState(false);
  const [tagsError, setTagsError] = useState<string | null>(null);
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [pendingTagIds, setPendingTagIds] = useState<number[]>([]);
  const [savingTags, setSavingTags] = useState(false);

  useEffect(() => {
    if (open) {
      setSelectedImageIndex(initialImageIndex);
    } else {
      setRecord(null);
      setError(null);
      setSelectedImageIndex(undefined);
    }
  }, [initialImageIndex, open]);

  useEffect(() => {
    if (!open || !recordId) {
      return;
    }

    let cancelled = false;
    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const detail = await fetchUsageRecordDetail(recordId);
        if (!cancelled) {
          setRecord(detail);
          setSelectedImageIndex((prev) => {
            if (prev !== undefined) {
              return prev;
            }
            return detail.output_images.length > 0 ? 0 : undefined;
          });
        }
      } catch (err) {
        if (!cancelled) {
          const message =
            err instanceof Error ? err.message : "加载记录详情失败";
          setError(message);
          setRecord(null);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void load();

    return () => {
      cancelled = true;
    };
  }, [open, recordId]);

  useEffect(() => {
    if (!open) {
      return;
    }

    let cancelled = false;
    const loadTags = async () => {
      setTagsLoading(true);
      setTagsError(null);
      try {
        const list = await fetchTags();
        if (!cancelled) {
          setTagOptions(list);
        }
      } catch (err) {
        if (!cancelled) {
          setTagsError(err instanceof Error ? err.message : "加载标签失败");
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
  }, [open]);

  useEffect(() => {
    if (!record) {
      setSelectedTagIds([]);
      setPendingTagIds([]);
      return;
    }
    const ids =
      record.tags?.map((tag) => tag.id).filter((id) => Number.isFinite(id)) ??
      [];
    setSelectedTagIds(ids);
    setPendingTagIds(ids);
  }, [record]);

  useEffect(() => {
    if (!record) {
      return;
    }
    if (record.output_images.length === 0) {
      if (selectedImageIndex !== undefined) {
        setSelectedImageIndex(undefined);
      }
      return;
    }
    if (selectedImageIndex === undefined) {
      setSelectedImageIndex(0);
      return;
    }
    if (!record.output_images[selectedImageIndex]) {
      setSelectedImageIndex(0);
    }
  }, [record, selectedImageIndex]);

  const handleRetry = async () => {
    if (!record || !onRetry) {
      return;
    }
    await onRetry(record);
  };

  const handleDelete = async () => {
    if (!record || !onDelete) {
      return;
    }
    await onDelete(record);
  };

  const mergedTagOptions = useMemo(() => {
    const map = new Map<number, Tag>();
    tagOptions.forEach((tag) => map.set(tag.id, tag));
    (record?.tags ?? []).forEach((tag) => {
      if (!map.has(tag.id)) {
        map.set(tag.id, tag);
      }
    });
    return Array.from(map.values()).sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [record?.tags, tagOptions]);

  const selectedTagOptions = useMemo(
    () => mergedTagOptions.filter((tag) => pendingTagIds.includes(tag.id)),
    [mergedTagOptions, pendingTagIds],
  );

  const hasTagChanges =
    selectedTagIds.length !== pendingTagIds.length ||
    selectedTagIds.some((id) => !pendingTagIds.includes(id)) ||
    pendingTagIds.some((id) => !selectedTagIds.includes(id));

  const canEditTags =
    Boolean(record) &&
    (isAdmin || (user && record?.user?.id && record.user.id === user.id));

  const handleSaveTags = async () => {
    if (!record || !canEditTags) {
      return;
    }
    setSavingTags(true);
    setTagsError(null);
    try {
      const updated = await updateUsageRecordTags(record.id, pendingTagIds);
      setRecord(updated);
      setSelectedTagIds(pendingTagIds);
      onTagsUpdated?.(updated);
    } catch (err) {
      setTagsError(err instanceof Error ? err.message : "更新标签失败");
    } finally {
      setSavingTags(false);
    }
  };

  const handleTagSelectionChange = (_: unknown, value: Tag[]) => {
    setTagsError(null);
    setPendingTagIds(value.map((tag) => tag.id));
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <Typography variant="h6">生成记录详情</Typography>
        <IconButton onClick={onClose} size="small">
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent dividers>
        {loading && (
          <Stack alignItems="center" spacing={1.5} sx={{ py: 4 }}>
            <CircularProgress />
            <Typography variant="body2" color="text.secondary">
              正在加载记录详情...
            </Typography>
          </Stack>
        )}

        {!loading && error && <Alert severity="error">{error}</Alert>}

        {!loading && !error && record && (
          <Stack spacing={3}>
            <Box>
              <Stack
                direction="row"
                spacing={1}
                alignItems="center"
                flexWrap="wrap"
                useFlexGap
              >
                <Typography variant="subtitle2" color="text.secondary">
                  关联标签
                </Typography>
                {(record.tags ?? []).length === 0 && (
                  <Chip label="未设置" size="small" variant="outlined" />
                )}
                {(record.tags ?? []).map((tag) => (
                  <Chip key={tag.id} label={tag.name} size="small" />
                ))}
              </Stack>

              {canEditTags && (
                <Stack spacing={1} sx={{ mt: 1.5 }}>
                  <Autocomplete
                    multiple
                    size="small"
                    options={mergedTagOptions}
                    value={selectedTagOptions}
                    onChange={handleTagSelectionChange}
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
                        label="选择或搜索标签"
                        placeholder="选择标签"
                      />
                    )}
                  />
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Button
                      variant="contained"
                      size="small"
                      disabled={!hasTagChanges || savingTags}
                      onClick={() => {
                        void handleSaveTags();
                      }}
                    >
                      {savingTags ? (
                        <CircularProgress size={16} color="inherit" />
                      ) : (
                        "保存标签"
                      )}
                    </Button>
                    <Button
                      variant="text"
                      size="small"
                      disabled={!hasTagChanges || savingTags}
                      onClick={() => setPendingTagIds(selectedTagIds)}
                    >
                      还原
                    </Button>
                  </Stack>
                  {tagsError && <Alert severity="warning">{tagsError}</Alert>}
                </Stack>
              )}
            </Box>

            {record.output_images.length > 0 && (
              <Box>
                <Typography
                  variant="subtitle2"
                  color="text.secondary"
                  gutterBottom
                >
                  输出媒体
                </Typography>
                <Stack
                  spacing={2}
                  direction={{ xs: "column", md: "row" }}
                  useFlexGap
                  flexWrap="wrap"
                  alignItems="stretch"
                >
                  {record.output_images.map((image, index) => {
                    const isSelected = selectedImageIndex === index;
                    const isPreparing = Boolean(
                      actionState?.preparingOutput &&
                        actionState?.preparingOutputIndex === index,
                    );
                    const canUse = Boolean(onUseOutputImage && image?.url);

                    const handlePreview = () => {
                      setSelectedImageIndex(index);
                      if (onPreviewOutputImage) {
                        onPreviewOutputImage(record, index);
                      }
                    };

                    return (
                      <Box
                        key={`${image.url}-${index}`}
                        sx={{
                          borderRadius: 2,
                          overflow: "hidden",
                          border: "2px solid",
                          borderColor: isSelected ? "primary.main" : "divider",
                          boxShadow: isSelected ? 4 : 0,
                          bgcolor: "background.paper",
                        }}
                      >
                        <ButtonBase
                          onClick={handlePreview}
                          sx={{
                            display: "block",
                            width: { xs: "100%", md: 360 },
                            bgcolor: "background.default",
                            "& img, & video": {
                              transition: "transform 0.3s ease",
                            },
                            "&:hover img, &:hover video": {
                              transform: "scale(1.01)",
                            },
                          }}
                        >
                          <Box
                            component={isVideoUrl(image.url) ? "video" : "img"}
                            src={image.url}
                            alt={`输出媒体 ${index + 1}`}
                            sx={{
                              width: "100%",
                              display: "block",
                              maxHeight: 420,
                              objectFit: "contain",
                              backgroundColor: "background.default",
                            }}
                            {...(isVideoUrl(image.url)
                              ? {
                                  controls: true,
                                  muted: true,
                                  playsInline: true,
                                }
                              : {})}
                          />
                        </ButtonBase>
                        <Stack
                          direction="row"
                          spacing={1}
                          justifyContent="flex-end"
                          sx={{ p: 1.5 }}
                        >
                          {canUse && (
                            <Button
                              size="small"
                              variant="outlined"
                              startIcon={
                                isPreparing ? (
                                  <CircularProgress size={16} color="inherit" />
                                ) : (
                                  <InputIcon />
                                )
                              }
                              disabled={isPreparing}
                              onClick={() => {
                                if (!onUseOutputImage || !image.url) {
                                  return;
                                }
                                void onUseOutputImage(record, image.url, index);
                              }}
                            >
                              带入生成
                            </Button>
                          )}
                        </Stack>
                      </Box>
                    );
                  })}
                </Stack>
              </Box>
            )}

            <Box>
              <Typography
                variant="subtitle2"
                color="text.secondary"
                gutterBottom
              >
                提示词
              </Typography>
              <Typography variant="body2" sx={{ whiteSpace: "pre-wrap" }}>
                {record.prompt || "（无提示词）"}
              </Typography>
            </Box>

            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
              <Chip label={`记录 #${record.id}`} />
              <Chip label={`厂商 ${record.provider_id}`} />
              <Chip label={`模型 ${record.model_id}`} />
              {record.size && <Chip label={`尺寸 ${record.size}`} />}
              <Chip label={formatDateTime(record.created_at)} />
              {record.user && (
                <Chip
                  label={`提交人 ${record.user.display_name || record.user.email}`}
                  color="primary"
                  variant="outlined"
                />
              )}
            </Stack>

            {record.output_text && (
              <Box>
                <Typography
                  variant="subtitle2"
                  color="text.secondary"
                  gutterBottom
                >
                  输出文本
                </Typography>
                <Typography variant="body2" sx={{ whiteSpace: "pre-wrap" }}>
                  {record.output_text}
                </Typography>
              </Box>
            )}

            {record.input_images.length > 0 && (
              <Box>
                <Typography
                  variant="subtitle2"
                  color="text.secondary"
                  gutterBottom
                >
                  输入媒体
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  {record.input_images.map((image, index) => (
                    <Box
                      key={`${image.url}-${index}`}
                      sx={{
                        width: 80,
                        height: 80,
                        borderRadius: 1,
                        overflow: "hidden",
                        border: "1px solid",
                        borderColor: "divider",
                      }}
                    >
                      <Box
                        component={isVideoUrl(image.url) ? "video" : "img"}
                        src={image.url}
                        alt={`输入媒体 ${index + 1}`}
                        sx={{
                          width: "100%",
                          height: "100%",
                          objectFit: "cover",
                          backgroundColor: "background.default",
                        }}
                        {...(isVideoUrl(image.url)
                          ? {
                              muted: true,
                              playsInline: true,
                              loop: true,
                            }
                          : {})}
                      />
                    </Box>
                  ))}
                </Stack>
              </Box>
            )}

            {record.error_message && (
              <Alert severity="warning">{record.error_message}</Alert>
            )}
          </Stack>
        )}
      </DialogContent>
      <DialogActions
        sx={{ justifyContent: "space-between", flexWrap: "wrap", gap: 1 }}
      >
        <Button onClick={onClose}>关闭</Button>
        <Stack
          direction={{ xs: "column", sm: "row" }}
          spacing={1}
          alignItems="stretch"
        >
          {onRetry && (
            <Button
              variant="contained"
              color="primary"
              startIcon={
                actionState?.retrying ? (
                  <CircularProgress size={16} color="inherit" />
                ) : (
                  <ReplayIcon />
                )
              }
              disabled={!record || actionState?.retrying}
              onClick={() => {
                void handleRetry();
              }}
            >
              再次生成
            </Button>
          )}

          {onDelete && (
            <Button
              variant="outlined"
              color="error"
              startIcon={
                actionState?.deleting ? (
                  <CircularProgress size={16} color="inherit" />
                ) : (
                  <DeleteOutlineIcon />
                )
              }
              disabled={!record || actionState?.deleting}
              onClick={() => {
                void handleDelete();
              }}
            >
              删除记录
            </Button>
          )}
        </Stack>
      </DialogActions>
    </Dialog>
  );
};

export default UsageRecordDetailDialog;
