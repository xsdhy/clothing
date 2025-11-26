import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  IconButton,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import SaveIcon from "@mui/icons-material/Save";
import AddIcon from "@mui/icons-material/Add";

import { createTag, deleteTag, fetchTags, updateTag } from "../ai";
import type { Tag } from "../types";

const TagManagementPage: React.FC = () => {
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newTagName, setNewTagName] = useState("");
  const [creating, setCreating] = useState(false);
  const [savingTagId, setSavingTagId] = useState<number | null>(null);
  const [deletingTagId, setDeletingTagId] = useState<number | null>(null);
  const [editValues, setEditValues] = useState<Record<number, string>>({});

  const fetchAllTags = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const list = await fetchTags();
      setTags(list);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载标签失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchAllTags();
  }, [fetchAllTags]);

  const handleCreate = async () => {
    const name = newTagName.trim();
    if (!name) {
      setError("请输入标签名称");
      return;
    }
    setCreating(true);
    setError(null);
    try {
      const created = await createTag(name);
      setTags((prev) => [...prev, created]);
      setNewTagName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建标签失败");
    } finally {
      setCreating(false);
    }
  };

  const handleUpdate = async (tag: Tag) => {
    const value = (editValues[tag.id] ?? tag.name ?? "").trim();
    if (!value || value === tag.name) {
      setEditValues((prev) => {
        const next = { ...prev };
        delete next[tag.id];
        return next;
      });
      return;
    }

    setSavingTagId(tag.id);
    setError(null);
    try {
      const updated = await updateTag(tag.id, value);
      setTags((prev) => prev.map((item) => (item.id === tag.id ? updated : item)));
      setEditValues((prev) => {
        const next = { ...prev };
        delete next[tag.id];
        return next;
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新标签失败");
    } finally {
      setSavingTagId(null);
    }
  };

  const handleDelete = async (tag: Tag) => {
    const usageCount = tag.usage_count ?? 0;
    const confirmed = window.confirm(
      usageCount > 0
        ? `该标签已关联 ${usageCount} 条记录，确认删除吗？`
        : "确认删除该标签吗？",
    );
    if (!confirmed) {
      return;
    }

    setDeletingTagId(tag.id);
    setError(null);
    try {
      await deleteTag(tag.id);
      setTags((prev) => prev.filter((item) => item.id !== tag.id));
      setEditValues((prev) => {
        const next = { ...prev };
        delete next[tag.id];
        return next;
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除标签失败");
    } finally {
      setDeletingTagId(null);
    }
  };

  const handleRefresh = () => {
    void fetchAllTags();
  };

  const isBusy = loading || creating || savingTagId !== null || deletingTagId !== null;

  const orderedTags = useMemo(
    () => [...tags].sort((a, b) => a.name.localeCompare(b.name)),
    [tags],
  );

  return (
    <Box sx={{ py: 6 }}>
      <Card elevation={1}>
        <CardHeader
          title="标签管理"
          subheader="为生成记录统一添加与维护标签，可用于筛选与分类。"
          action={
            <Tooltip title="刷新标签列表">
              <span>
                <IconButton onClick={handleRefresh} disabled={loading}>
                  {loading ? <CircularProgress size={18} /> : <RefreshIcon />}
                </IconButton>
              </span>
            </Tooltip>
          }
        />
        <CardContent>
          <Stack spacing={3}>
            {error && <Alert severity="warning">{error}</Alert>}

            <Stack
              direction={{ xs: "column", md: "row" }}
              spacing={2}
              alignItems={{ xs: "stretch", md: "center" }}
            >
              <TextField
                label="新建标签"
                placeholder="输入标签名称"
                value={newTagName}
                onChange={(event) => setNewTagName(event.target.value)}
                size="small"
                sx={{ minWidth: 260, maxWidth: 420 }}
                disabled={creating}
              />
              <Button
                variant="contained"
                startIcon={
                  creating ? <CircularProgress size={16} color="inherit" /> : <AddIcon />
                }
                onClick={() => {
                  void handleCreate();
                }}
                disabled={creating}
              >
                创建标签
              </Button>
            </Stack>

            {loading ? (
              <Stack alignItems="center" spacing={1.5} sx={{ py: 4 }}>
                <CircularProgress />
                <Typography variant="body2" color="text.secondary">
                  正在加载标签...
                </Typography>
              </Stack>
            ) : orderedTags.length === 0 ? (
              <Alert severity="info">暂无标签，创建后可用于筛选生成记录。</Alert>
            ) : (
              <Stack spacing={1.5}>
                {orderedTags.map((tag) => {
                  const editingValue = editValues[tag.id] ?? tag.name ?? "";
                  const isSaving = savingTagId === tag.id;
                  const isDeleting = deletingTagId === tag.id;
                  const disableActions = isBusy && !isSaving && !isDeleting;
                  const usageCount = tag.usage_count ?? 0;

                  return (
                    <Box
                      key={tag.id}
                      sx={{
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 2,
                        p: 2,
                        bgcolor: "background.paper",
                      }}
                    >
                      <Stack
                        direction={{ xs: "column", md: "row" }}
                        spacing={2}
                        alignItems={{ xs: "stretch", md: "center" }}
                        justifyContent="space-between"
                      >
                        <Box sx={{ flex: 1, minWidth: 240 }}>
                          <TextField
                            fullWidth
                            size="small"
                            label="标签名称"
                            value={editingValue}
                            onChange={(event) =>
                              setEditValues((prev) => ({
                                ...prev,
                                [tag.id]: event.target.value,
                              }))
                            }
                            disabled={isSaving || isDeleting}
                          />
                          <Typography variant="caption" color="text.secondary">
                            已关联 {usageCount} 条生成记录
                          </Typography>
                        </Box>
                        <Stack direction="row" spacing={1}>
                          <Button
                            variant="outlined"
                            startIcon={
                              isSaving ? (
                                <CircularProgress size={14} color="inherit" />
                              ) : (
                                <SaveIcon fontSize="small" />
                              )
                            }
                            onClick={() => {
                              void handleUpdate(tag);
                            }}
                            disabled={
                              isSaving ||
                              isDeleting ||
                              disableActions ||
                              editingValue.trim() === "" ||
                              editingValue.trim() === tag.name
                            }
                          >
                            保存
                          </Button>
                          <Button
                            color="error"
                            variant="text"
                            startIcon={
                              isDeleting ? (
                                <CircularProgress size={14} color="inherit" />
                              ) : (
                                <DeleteOutlineIcon fontSize="small" />
                              )
                            }
                            onClick={() => {
                              void handleDelete(tag);
                            }}
                            disabled={isSaving || isDeleting || disableActions}
                          >
                            删除
                          </Button>
                        </Stack>
                      </Stack>
                    </Box>
                  );
                })}
              </Stack>
            )}
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};

export default TagManagementPage;
