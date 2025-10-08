import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  ButtonBase,
  Card,
  CardActions,
  CardContent,
  Chip,
  CircularProgress,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Typography,
} from '@mui/material';
import OpenInFullIcon from '@mui/icons-material/OpenInFull';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import ReplayIcon from '@mui/icons-material/Replay';
import InputIcon from '@mui/icons-material/Input';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import RefreshIcon from '@mui/icons-material/Refresh';
import CloseIcon from '@mui/icons-material/Close';
import { useNavigate } from 'react-router-dom';

import { deleteUsageRecord, fetchUsageRecords } from '../ai';
import type { UsageRecord } from '../types';
import ImageViewer, { type ImageViewerItem } from '../components/ImageViewer';

const PAGE_SIZE = 20;

interface GalleryItem {
  recordId: number;
  imageIndex: number;
  url: string;
  prompt: string;
  createdAt: string;
  providerId: string;
  modelId: string;
}

interface SelectedDetail {
  recordId: number;
  imageIndex?: number;
}

const formatDateTime = (value: string): string => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};

const GeneratedImageGalleryPage: React.FC = () => {
  const navigate = useNavigate();
  const [records, setRecords] = useState<UsageRecord[]>([]);
  const [initialLoading, setInitialLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [nextPage, setNextPage] = useState(1);
  const [metaTotal, setMetaTotal] = useState(0);
  const [selectedDetail, setSelectedDetail] = useState<SelectedDetail | null>(null);
  const [viewerState, setViewerState] = useState<{ open: boolean; index: number }>({ open: false, index: 0 });
  const [retryingRecordId, setRetryingRecordId] = useState<number | null>(null);
  const [preparingOutputAsInput, setPreparingOutputAsInput] = useState<{ recordId: number; index: number } | null>(null);
  const [deletingRecordId, setDeletingRecordId] = useState<number | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const recordMap = useMemo(() => new Map(records.map((record) => [record.id, record])), [records]);

  const galleryItems = useMemo<GalleryItem[]>(() => {
    const items: GalleryItem[] = [];
    records.forEach((record) => {
      (record.output_images ?? []).forEach((image, index) => {
        if (!image?.url) {
          return;
        }
        items.push({
          recordId: record.id,
          imageIndex: index,
          url: image.url,
          prompt: record.prompt,
          createdAt: record.created_at,
          providerId: record.provider_id,
          modelId: record.model_id,
        });
      });
    });
    return items;
  }, [records]);

  const viewerItems = useMemo<ImageViewerItem[]>(
    () =>
      galleryItems.map((item) => ({
        src: item.url,
        key: `${item.recordId}-${item.imageIndex}-${item.url}`,
        title: `记录 #${item.recordId} 第 ${item.imageIndex + 1} 张`,
        downloadName: `record-${item.recordId}-image-${item.imageIndex + 1}.png`,
      })),
    [galleryItems]
  );

  useEffect(() => {
    setViewerState((prev) => {
      if (!prev.open) {
        return prev;
      }
      if (viewerItems.length === 0) {
        return { open: false, index: 0 };
      }
      const nextIndex = Math.min(prev.index, Math.max(viewerItems.length - 1, 0));
      if (nextIndex !== prev.index) {
        return { ...prev, index: nextIndex };
      }
      return prev;
    });
  }, [viewerItems.length]);

  const loadRecords = useCallback(
    async (pageToLoad: number, append: boolean) => {
      if (append) {
        setLoadingMore(true);
        setLoadMoreError(null);
      } else {
        setInitialLoading(true);
        setError(null);
      }

      try {
        const result = await fetchUsageRecords(pageToLoad, PAGE_SIZE, { result: 'success' });

        let mergedRecords: UsageRecord[] = [];
        setRecords((prev) => {
          if (!append) {
            mergedRecords = result.records;
            return result.records;
          }

          const existingIndex = new Map(prev.map((record, index) => [record.id, index]));
          const next = [...prev];

          result.records.forEach((record) => {
            const index = existingIndex.get(record.id);
            if (index !== undefined) {
              next[index] = record;
            } else {
              next.push(record);
            }
          });

          mergedRecords = next;
          return next;
        });

        const meta = result.meta;
        const total = meta?.total ?? mergedRecords.length;
        const page = meta?.page ?? pageToLoad;
        const pageSize = meta?.page_size ?? PAGE_SIZE;

        setMetaTotal(total);
        setHasMore(page * pageSize < total);
        setNextPage(page + 1);
      } catch (err) {
        const message = err instanceof Error ? err.message : '加载图片失败';
        if (append) {
          setLoadMoreError(message);
        } else {
          setError(message);
        }
      } finally {
        if (append) {
          setLoadingMore(false);
        } else {
          setInitialLoading(false);
        }
      }
    },
    []
  );

  useEffect(() => {
    void loadRecords(1, false);
  }, [loadRecords]);

  const handleRefresh = useCallback(() => {
    setNextPage(1);
    setHasMore(true);
    void loadRecords(1, false);
  }, [loadRecords]);

  const handleLoadMore = useCallback(() => {
    if (!hasMore || loadingMore || initialLoading) {
      return;
    }
    void loadRecords(nextPage, true);
  }, [hasMore, initialLoading, loadRecords, loadingMore, nextPage]);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasMore) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            handleLoadMore();
          }
        });
      },
      { root: null, rootMargin: '200px 0px', threshold: 0 }
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [handleLoadMore, hasMore]);

  useEffect(() => {
    if (!selectedDetail) {
      return;
    }
    if (!recordMap.has(selectedDetail.recordId)) {
      setSelectedDetail(null);
    }
  }, [recordMap, selectedDetail]);

  const loadImageAsDataUrl = useCallback(async (url: string): Promise<string> => {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`无法读取图片: ${response.status} ${response.statusText}`);
    }

    const blob = await response.blob();

    return await new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result;
        if (typeof result === 'string') {
          resolve(result);
        } else {
          reject(new Error('图片转换失败'));
        }
      };
      reader.onerror = () => {
        reject(new Error('图片读取失败'));
      };
      reader.readAsDataURL(blob);
    });
  }, []);

  const handleRetry = useCallback(
    async (record: UsageRecord) => {
      try {
        setRetryingRecordId(record.id);
        const base64Images = await Promise.all(record.input_images.map((image) => loadImageAsDataUrl(image.url)));

        const state = {
          prompt: record.prompt,
          inputImages: base64Images,
          providerId: record.provider_id,
          modelId: record.model_id,
          size: record.size,
        } as const;

        navigate('/custom', { state });
      } catch (err) {
        const message = err instanceof Error ? err.message : '重新创建生成任务失败';
        setError(message);
      } finally {
        setRetryingRecordId(null);
      }
    },
    [loadImageAsDataUrl, navigate]
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

        navigate('/custom', { state });
      } catch (err) {
        const message = err instanceof Error ? err.message : '无法将图片带入生成页面';
        setError(message);
      } finally {
        setPreparingOutputAsInput(null);
      }
    },
    [loadImageAsDataUrl, navigate]
  );

  const handleDeleteRecord = useCallback(
    async (record: UsageRecord) => {
      if (!window.confirm('确认删除该生成记录吗？')) {
        return;
      }

      try {
        setDeletingRecordId(record.id);
        await deleteUsageRecord(record.id);

        let nextLength = 0;
        setRecords((prev) => {
          const next = prev.filter((item) => item.id !== record.id);
          nextLength = next.length;
          return next;
        });

        setMetaTotal((prevTotal) => {
          const nextTotal = Math.max(0, prevTotal - 1);
          setHasMore(nextLength < nextTotal);
          return nextTotal;
        });

        setSelectedDetail((prev) => {
          if (prev?.recordId === record.id) {
            return null;
          }
          return prev;
        });
      } catch (err) {
        const message = err instanceof Error ? err.message : '删除生成记录失败';
        setError(message);
      } finally {
        setDeletingRecordId(null);
      }
    },
    []
  );

  const handleOpenViewer = useCallback((index: number) => {
    setViewerState({ open: true, index });
  }, []);

  const handleCloseViewer = useCallback(() => {
    setViewerState((prev) => ({ ...prev, open: false }));
  }, []);

  const handleDownloadViewerImage = useCallback((_index: number, item: ImageViewerItem) => {
    if (!item?.src) {
      return;
    }

    const link = document.createElement('a');
    link.href = item.src;
    link.download = item.downloadName ?? 'generated-image.png';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }, []);

  const handleOpenDetails = useCallback((recordId: number, imageIndex?: number) => {
    if (!recordMap.has(recordId)) {
      return;
    }
    setSelectedDetail({ recordId, imageIndex });
  }, [recordMap]);

  const handleCloseDetails = useCallback(() => {
    setSelectedDetail(null);
  }, []);

  const selectedRecord = selectedDetail ? recordMap.get(selectedDetail.recordId) ?? null : null;
  const selectedImage = selectedRecord && selectedDetail?.imageIndex !== undefined
    ? selectedRecord.output_images[selectedDetail.imageIndex] ?? null
    : null;

  return (
    <Box sx={{ py: 6 }}>
      <Container maxWidth="xl">
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ xs: 'flex-start', sm: 'center' }} justifyContent="space-between" sx={{ mb: 4 }}>
          <Box>
            <Typography variant="h4" gutterBottom>
              生成图片库
            </Typography>
            <Typography variant="body2" color="text.secondary">
              浏览所有生成的图片，支持瀑布流查看、放大预览、下拉加载与关联记录操作。
            </Typography>
          </Box>
          <Button variant="contained" color="primary" startIcon={<RefreshIcon />} onClick={handleRefresh} disabled={initialLoading}>
            刷新
          </Button>
        </Stack>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }} action={
            <Button color="inherit" size="small" onClick={handleRefresh}>
              重试
            </Button>
          }>
            {error}
          </Alert>
        )}

        {initialLoading ? (
          <Stack alignItems="center" sx={{ py: 6 }}>
            <CircularProgress />
            <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
              正在加载生成图片...
            </Typography>
          </Stack>
        ) : galleryItems.length === 0 ? (
          <Alert severity="info">暂无可展示的生成图片，尝试先进行一次图片生成。</Alert>
        ) : (
          <Box
            sx={{
              columnCount: { xs: 1, sm: 2, md: 3, lg: 4 },
              columnGap: 2,
            }}
          >
            {galleryItems.map((item, index) => (
              <Box key={`${item.recordId}-${item.imageIndex}-${item.url}`} sx={{ breakInside: 'avoid', mb: 2 }}>
                <Card elevation={1} sx={{ overflow: 'hidden' }}>
                  <ButtonBase
                    onClick={() => handleOpenViewer(index)}
                    sx={{
                      display: 'block',
                      width: '100%',
                      '& img': {
                        transition: 'transform 0.3s ease',
                      },
                      '&:hover img': {
                        transform: 'scale(1.02)',
                      },
                    }}
                  >
                    <Box component="img" src={item.url} alt={item.prompt || `记录 ${item.recordId} 生成图片`}
                      sx={{ width: '100%', display: 'block' }}
                    />
                  </ButtonBase>
                  <CardContent>
                    <Stack spacing={1}>
                      <Typography variant="subtitle2" title={item.prompt} sx={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {item.prompt || '（无提示词）'}
                      </Typography>
                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                        <Chip size="small" label={`记录 #${item.recordId}`} />
                        <Chip size="small" label={item.providerId} />
                        <Chip size="small" label={item.modelId} />
                      </Stack>
                      <Typography variant="caption" color="text.secondary">
                        {formatDateTime(item.createdAt)}
                      </Typography>
                    </Stack>
                  </CardContent>
                  <CardActions sx={{ justifyContent: 'space-between', flexWrap: 'wrap' }}>
                    <Button size="small" startIcon={<OpenInFullIcon fontSize="small" />} onClick={() => handleOpenViewer(index)}>
                      放大查看
                    </Button>
                    <Button size="small" startIcon={<InfoOutlinedIcon fontSize="small" />} onClick={() => handleOpenDetails(item.recordId, item.imageIndex)}>
                      查看详情
                    </Button>
                  </CardActions>
                </Card>
              </Box>
            ))}
          </Box>
        )}

        <Box ref={sentinelRef} sx={{ height: 1 }} />

        {loadingMore && (
          <Stack direction="row" spacing={1} alignItems="center" justifyContent="center" sx={{ py: 3 }}>
            <CircularProgress size={20} />
            <Typography variant="body2" color="text.secondary">
              正在加载更多图片...
            </Typography>
          </Stack>
        )}

        {loadMoreError && (
          <Alert
            severity="warning"
            sx={{ mt: 2 }}
            action={
              <Button color="inherit" size="small" onClick={() => void loadRecords(nextPage, true)}>
                重试
              </Button>
            }
          >
            {loadMoreError}
          </Alert>
        )}

        {!hasMore && galleryItems.length > 0 && (
          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', textAlign: 'center', py: 2 }}>
            已加载全部 {metaTotal} 张生成图片
          </Typography>
        )}

        <ImageViewer
          open={viewerState.open}
          onClose={handleCloseViewer}
          images={viewerItems}
          initialIndex={viewerState.index}
          showDownload
          onDownload={handleDownloadViewerImage}
        />

        <Dialog open={Boolean(selectedRecord)} onClose={handleCloseDetails} maxWidth="md" fullWidth>
          <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Typography variant="h6">生成记录详情</Typography>
            <IconButton onClick={handleCloseDetails} size="small">
              <CloseIcon />
            </IconButton>
          </DialogTitle>
          <DialogContent dividers>
            {selectedRecord && (
              <Stack spacing={3}>
                {selectedImage && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      当前选中图片
                    </Typography>
                    <ButtonBase
                      onClick={() => {
                        const index = galleryItems.findIndex(
                          (item) => item.recordId === selectedRecord.id && item.imageIndex === (selectedDetail?.imageIndex ?? -1)
                        );
                        if (index !== -1) {
                          handleOpenViewer(index);
                        }
                      }}
                      sx={{
                        borderRadius: 1,
                        overflow: 'hidden',
                        border: '1px solid',
                        borderColor: 'divider',
                        display: 'inline-block',
                      }}
                    >
                      <Box component="img" src={selectedImage.url} alt={selectedRecord.prompt}
                        sx={{ width: { xs: 200, sm: 260 }, display: 'block' }}
                      />
                    </ButtonBase>
                  </Box>
                )}

                <Box>
                  <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                    提示词
                  </Typography>
                  <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                    {selectedRecord.prompt || '（无提示词）'}
                  </Typography>
                </Box>

                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  <Chip label={`记录 #${selectedRecord.id}`} />
                  <Chip label={`厂商 ${selectedRecord.provider_id}`} />
                  <Chip label={`模型 ${selectedRecord.model_id}`} />
                  {selectedRecord.size && <Chip label={`尺寸 ${selectedRecord.size}`} />}
                  <Chip label={formatDateTime(selectedRecord.created_at)} />
                </Stack>

                {selectedRecord.output_text && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      输出文本
                    </Typography>
                    <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                      {selectedRecord.output_text}
                    </Typography>
                  </Box>
                )}

                {selectedRecord.input_images.length > 0 && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      输入图片
                    </Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                      {selectedRecord.input_images.map((image, index) => (
                        <Box key={`${image.url}-${index}`} sx={{ width: 80, height: 80, borderRadius: 1, overflow: 'hidden', border: '1px solid', borderColor: 'divider' }}>
                          <Box component="img" src={image.url} alt={`输入图片 ${index + 1}`}
                            sx={{ width: '100%', height: '100%', objectFit: 'cover' }}
                          />
                        </Box>
                      ))}
                    </Stack>
                  </Box>
                )}

                {selectedRecord.error_message && (
                  <Alert severity="warning">{selectedRecord.error_message}</Alert>
                )}
              </Stack>
            )}
          </DialogContent>
          {selectedRecord && (
            <DialogActions sx={{ justifyContent: 'space-between', flexWrap: 'wrap', gap: 1 }}>
              <Button onClick={handleCloseDetails}>关闭</Button>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems="stretch">
                <Button
                  variant="outlined"
                  startIcon={
                    preparingOutputAsInput && preparingOutputAsInput.recordId === selectedRecord.id && selectedDetail?.imageIndex !== undefined &&
                    preparingOutputAsInput.index === selectedDetail.imageIndex ? (
                      <CircularProgress size={16} color="inherit" />
                    ) : (
                      <InputIcon />
                    )
                  }
                  disabled={
                    selectedDetail?.imageIndex === undefined ||
                    preparingOutputAsInput?.recordId === selectedRecord.id && preparingOutputAsInput.index === selectedDetail.imageIndex
                  }
                  onClick={() => {
                    if (selectedDetail?.imageIndex === undefined) {
                      return;
                    }
                    const image = selectedRecord.output_images[selectedDetail.imageIndex];
                    if (!image) {
                      return;
                    }
                    void handleUseOutputImage(selectedRecord, image.url, selectedDetail.imageIndex);
                  }}
                >
                  带入生成
                </Button>
                <Button
                  variant="contained"
                  color="primary"
                  startIcon={retryingRecordId === selectedRecord.id ? <CircularProgress size={16} color="inherit" /> : <ReplayIcon />}
                  disabled={retryingRecordId === selectedRecord.id}
                  onClick={() => void handleRetry(selectedRecord)}
                >
                  再次生成
                </Button>
                <Button
                  variant="outlined"
                  color="error"
                  startIcon={deletingRecordId === selectedRecord.id ? <CircularProgress size={16} color="inherit" /> : <DeleteOutlineIcon />}
                  disabled={deletingRecordId === selectedRecord.id}
                  onClick={() => void handleDeleteRecord(selectedRecord)}
                >
                  删除记录
                </Button>
              </Stack>
            </DialogActions>
          )}
        </Dialog>
      </Container>
    </Box>
  );
};

export default GeneratedImageGalleryPage;
