import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Box, Button, ButtonBase, Card, CardActions, CircularProgress, Container, Stack, Typography } from '@mui/material';
import OpenInFullIcon from '@mui/icons-material/OpenInFull';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import RefreshIcon from '@mui/icons-material/Refresh';
import { useNavigate } from 'react-router-dom';

import { deleteUsageRecord, fetchUsageRecords } from '../ai';
import type { UsageRecord } from '../types';
import ImageViewer, { type ImageViewerItem } from '../components/ImageViewer';
import UsageRecordDetailDialog from '../components/UsageRecordDetailDialog';

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

  const selectedRecordId = selectedDetail?.recordId ?? null;
  const preparingMatch =
    selectedRecordId !== null && preparingOutputAsInput?.recordId === selectedRecordId;
  const preparingImageIndex = preparingMatch ? preparingOutputAsInput?.index : undefined;

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
 
                  <CardActions sx={{ justifyContent: 'space-between', flexWrap: 'wrap' }}>
                    <Button size="small" startIcon={<InfoOutlinedIcon fontSize="small" />} onClick={() => handleOpenDetails(item.recordId, item.imageIndex)}>
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

        <UsageRecordDetailDialog
          open={Boolean(selectedDetail)}
          recordId={selectedDetail?.recordId ?? null}
          initialImageIndex={selectedDetail?.imageIndex}
          onClose={handleCloseDetails}
          onRetry={(record) => {
            void handleRetry(record);
          }}
          onDelete={(record) => {
            void handleDeleteRecord(record);
          }}
          onUseOutputImage={(record, imageUrl, imageIndex) => {
            void handleUseOutputImage(record, imageUrl, imageIndex);
          }}
          onPreviewOutputImage={(record, imageIndex) => {
            const index = galleryItems.findIndex(
              (item) => item.recordId === record.id && item.imageIndex === imageIndex
            );
            if (index !== -1) {
              handleOpenViewer(index);
            }
          }}
          actionState={{
            retrying: selectedRecordId !== null && retryingRecordId === selectedRecordId,
            deleting: selectedRecordId !== null && deletingRecordId === selectedRecordId,
            preparingOutput: typeof preparingImageIndex === 'number',
            preparingOutputIndex: typeof preparingImageIndex === 'number' ? preparingImageIndex : undefined,
          }}
        />
      </Container>
    </Box>
  );
};

export default GeneratedImageGalleryPage;
