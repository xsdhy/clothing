import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  ButtonBase,
  Card,
  CardContent,
  CardHeader,
  Chip,
  CircularProgress,
  Container,
  Divider,
  Grid,
  IconButton,
  Collapse,
  Pagination,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material';

import ReplayIcon from '@mui/icons-material/Replay';
import AddPhotoAlternateIcon from '@mui/icons-material/AddPhotoAlternate';
import { useNavigate } from 'react-router-dom';

import { fetchUsageRecords, type UsageRecordsResult } from '../ai';
import type { UsageRecord } from '../types';
import ImageViewer from '../components/ImageViewer';

const PAGE_SIZE = 10;

const GenerationHistoryPage: React.FC = () => {
  const [records, setRecords] = useState<UsageRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState<UsageRecordsResult['meta'] | null>(null);
  const [previewImage, setPreviewImage] = useState<{ url: string; alt: string } | null>(null);
  const [retryingRecordId, setRetryingRecordId] = useState<number | null>(null);
  const [preparingOutputAsInput, setPreparingOutputAsInput] = useState<{ recordId: number; index: number } | null>(null);
  const [expandedRecords, setExpandedRecords] = useState<Record<number, boolean>>({});
  const navigate = useNavigate();

  const totalPages = useMemo(() => {
    if (!meta || meta.page_size <= 0) {
      return 1;
    }
    return Math.max(1, Math.ceil(meta.total / meta.page_size));
  }, [meta]);

  const loadRecords = useCallback(async (targetPage: number) => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchUsageRecords(targetPage, PAGE_SIZE);
      setRecords(result.records);
      setMeta(result.meta);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取记录失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadRecords(page);
  }, [loadRecords, page]);

  const handleRefresh = useCallback(() => {
    void loadRecords(page);
  }, [loadRecords, page]);

  const handlePageChange = useCallback((_event: React.ChangeEvent<unknown>, value: number) => {
    setPage(value);
  }, []);

  const handleImageClick = useCallback((url: string, alt: string) => {
    setPreviewImage({ url, alt });
  }, []);

  const handleClosePreview = useCallback(() => {
    setPreviewImage(null);
  }, []);

  const handleDownloadPreview = useCallback(() => {
    if (!previewImage) {
      return;
    }

    const link = document.createElement('a');
    link.href = previewImage.url;
    link.download = previewImage.alt.replace(/\s+/g, '_');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }, [previewImage]);

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

  const handleToggleDetails = useCallback((recordId: number) => {
    setExpandedRecords((prev) => ({
      ...prev,
      [recordId]: !prev[recordId],
    }));
  }, []);

  return (
    <Box sx={{ py: 6 }}>
      <Container maxWidth="lg">
        <Stack direction={{ xs: 'column', sm: 'row' }} alignItems={{ xs: 'flex-start', sm: 'center' }} justifyContent="space-between" spacing={2} sx={{ mb: 4 }}>
          <Box>
            <Typography variant="h4" gutterBottom>
              生成记录
            </Typography>
            <Typography variant="body2" color="text.secondary">
              查看每次调用模型的输入提示、生成结果与可能的错误信息。
            </Typography>
          </Box>
          <Button variant="contained" color="primary" onClick={handleRefresh} disabled={loading}>
            刷新
          </Button>
        </Stack>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        {loading && (
          <Stack direction="row" spacing={2} alignItems="center" justifyContent="center" sx={{ my: 6 }}>
            <CircularProgress size={28} />
            <Typography variant="body2" color="text.secondary">
              正在加载生成记录…
            </Typography>
          </Stack>
        )}

        {!loading && records.length === 0 && !error && (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="h6" gutterBottom>
              暂无生成记录
            </Typography>
            <Typography variant="body2" color="text.secondary">
              完成一次图片生成后，可在此处查看完整的调用详情。
            </Typography>
          </Box>
        )}

        <Grid container spacing={3} sx={{ opacity: loading ? 0.7 : 1 }}>
          {records.map((record) => {
            const createdAt = new Date(record.created_at).toLocaleString();
            const hasError = Boolean(record.error_message);

            return (
              <Grid item xs={12} key={`${record.id}-${record.created_at}`}>
                <Card elevation={1}>
                  <CardHeader
                    title={`${record.provider_id}/${record.model_id}`}
                    action={
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Button variant="text" size="small" onClick={() => handleToggleDetails(record.id)}>
                          {expandedRecords[record.id] ? '收起详情' : '查看详情'}
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
                        <Chip
                          color={hasError ? 'error' : 'success'}
                          variant="outlined"
                          label={hasError ? '生成异常' : '生成成功'}
                        />
                      </Stack>
                    }
                  />
                  <CardContent>
                    <Stack spacing={2}>
                      {(record.input_images.length > 0 || record.output_images.length > 0) && (
                        <Box>
                          {/* <Typography variant="subtitle2" color="text.primary">
                            输入 / 输出图片
                          </Typography> */}
                          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
                            {record.input_images.map((image, index) => {
                              const alt = `输入图片 ${index + 1}`;
                              return (
                                <ButtonBase
                                  key={`${record.id}-in-${index}`}
                                  onClick={() => handleImageClick(image.url, alt)}
                                  sx={{
                                    width: 140,
                                    height: 140,
                                    borderRadius: 1,
                                    overflow: 'hidden',
                                    border: '1px solid',
                                    borderColor: 'divider',
                                    bgcolor: 'background.paper',
                                    position: 'relative',
                                  }}
                                >
                                  <Box
                                    component="img"
                                    src={image.url}
                                    alt={alt}
                                    loading="lazy"
                                    sx={{ width: '100%', height: '100%', objectFit: 'cover' }}
                                  />
                                  <Chip
                                    label={`${index + 1}`}
                                    size="small"
                                    color="default"
                                    sx={{
                                      position: 'absolute',
                                      top: 6,
                                      left: 6,
                                      bgcolor: 'rgba(0,0,0,0.6)',
                                      color: 'common.white',
                                    }}
                                  />
                                </ButtonBase>
                              );
                            })}
                            {record.output_images.map((image, index) => {
                              const alt = `输出图片 ${index + 1}`;
                              return (
                                <ButtonBase
                                  key={`${record.id}-out-${index}`}
                                  onClick={() => handleImageClick(image.url, alt)}
                                  sx={{
                                    width: 140,
                                    height: 140,
                                    borderRadius: 1,
                                    overflow: 'hidden',
                                    border: '1px solid',
                                    borderColor: 'divider',
                                    bgcolor: 'background.paper',
                                    position: 'relative',
                                  }}
                                >
                                  <Box
                                    component="img"
                                    src={image.url}
                                    alt={alt}
                                    loading="lazy"
                                    sx={{ width: '100%', height: '100%', objectFit: 'cover' }}
                                  />
                                  <Chip
                                    label={`${index + 1}`}
                                    size="small"
                                    color="primary"
                                    sx={{
                                      position: 'absolute',
                                      top: 6,
                                      left: 6,
                                    }}
                                  />
                                  <Tooltip title="作为输入图片使用">
                                    <span>
                                      <IconButton
                                        size="small"
                                        color="primary"
                                        sx={{
                                          position: 'absolute',
                                          top: 6,
                                          right: 6,
                                          bgcolor: 'rgba(255,255,255,0.85)',
                                          '&:hover': {
                                            bgcolor: 'rgba(255,255,255,0.95)',
                                          },
                                        }}
                                        onClick={(event) => {
                                          event.stopPropagation();
                                          void handleUseOutputImage(record, image.url, index);
                                        }}
                                        disabled={
                                          preparingOutputAsInput?.recordId === record.id &&
                                          preparingOutputAsInput?.index === index
                                        }
                                      >
                                        {preparingOutputAsInput?.recordId === record.id &&
                                        preparingOutputAsInput?.index === index ? (
                                          <CircularProgress size={16} color="inherit" />
                                        ) : (
                                          <AddPhotoAlternateIcon fontSize="small" />
                                        )}
                                      </IconButton>
                                    </span>
                                  </Tooltip>
                                </ButtonBase>
                              );
                            })}
                          </Box>
                        </Box>
                      )}

                      <Collapse in={expandedRecords[record.id] ?? false} timeout="auto" unmountOnExit>
                        <Stack spacing={1.5} sx={{ mt: 1 }}>
                          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ xs: 'flex-start', sm: 'center' }}>
                            <Chip label={`供应商：${record.provider_id}`} color="primary" variant="outlined" size="small" />
                            <Chip label={`模型：${record.model_id}`} color="primary" variant="outlined" size="small" />
                            {record.size && <Chip label={`尺寸：${record.size}`} variant="outlined" size="small" />}
                          </Stack>

                            {record.prompt && (
                            <Box>
                              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                                输入文本
                              </Typography>
                              <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                                {record.prompt || '（无提示词）'}
                              </Typography>
                            </Box>
                          )}

                          {record.output_text && (
                            <Box>
                              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                                输出文本
                              </Typography>
                              <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                                {record.output_text}
                              </Typography>
                            </Box>
                          )}

                          {hasError && (
                            <Alert severity="warning">{record.error_message}</Alert>
                          )}

                          <Divider sx={{ my: 1 }} />

                          <Typography variant="caption" color="text.secondary">
                            记录 ID：{record.id} - {createdAt}
                          </Typography>
                        </Stack>
                      </Collapse>

                    </Stack>
                  </CardContent>
                </Card>
              </Grid>
            );
          })}
        </Grid>

        {totalPages > 1 && (
          <Stack alignItems="center" sx={{ mt: 4 }}>
            <Pagination count={totalPages} page={Math.min(page, totalPages)} onChange={handlePageChange} color="primary" />
            {meta && (
              <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
                共 {meta.total} 条记录，每页 {meta.page_size} 条
              </Typography>
            )}
          </Stack>
        )}

        <ImageViewer
          open={Boolean(previewImage)}
          onClose={handleClosePreview}
          imageUrl={previewImage?.url ?? ''}
          title={previewImage?.alt}
          showDownload
          onDownload={handleDownloadPreview}
        />
      </Container>
    </Box>
  );
};

export default GenerationHistoryPage;
