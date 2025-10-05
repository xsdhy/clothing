import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Chip,
  CircularProgress,
  Container,
  Divider,
  Grid,
  Pagination,
  Stack,
  Typography,
} from '@mui/material';

import { fetchUsageRecords, type UsageRecordsResult } from '../ai';
import type { UsageRecord } from '../types';

const PAGE_SIZE = 10;

const GenerationHistoryPage: React.FC = () => {
  const [records, setRecords] = useState<UsageRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState<UsageRecordsResult['meta'] | null>(null);

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
                    title={
                      <Typography variant="h6" component="div" sx={{ wordBreak: 'break-word' }}>
                        {record.prompt || '（无提示词）'}
                      </Typography>
                    }
                    subheader={`创建于 ${createdAt}`}
                    action={
                      <Chip
                        color={hasError ? 'error' : 'success'}
                        variant="outlined"
                        label={hasError ? '生成异常' : '生成成功'}
                      />
                    }
                  />
                  <CardContent>
                    <Stack spacing={2}>
                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ xs: 'flex-start', sm: 'center' }}>
                        <Chip label={`供应商：${record.provider_id}`} color="primary" variant="outlined" size="small" />
                        <Chip label={`模型：${record.model_id}`} color="primary" variant="outlined" size="small" />
                        {record.size && <Chip label={`尺寸：${record.size}`} variant="outlined" size="small" />}
                      </Stack>

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

                      {record.input_images.length > 0 && (
                        <Box>
                          <Typography variant="subtitle2" color="text.secondary">
                            输入图片
                          </Typography>
                          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
                            {record.input_images.map((image, index) => (
                              <Box
                                component="img"
                                key={`${record.id}-in-${index}`}
                                src={image.url}
                                alt={`输入图片 ${index + 1}`}
                                loading="lazy"
                                sx={{
                                  width: 96,
                                  height: 96,
                                  objectFit: 'cover',
                                  borderRadius: 1,
                                  border: '1px solid',
                                  borderColor: 'divider',
                                  bgcolor: 'background.paper',
                                }}
                              />
                            ))}
                          </Box>
                        </Box>
                      )}

                      {record.output_images.length > 0 && (
                        <Box>
                          <Typography variant="subtitle2" color="text.secondary">
                            输出图片
                          </Typography>
                          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
                            {record.output_images.map((image, index) => (
                              <Box
                                component="img"
                                key={`${record.id}-out-${index}`}
                                src={image.url}
                                alt={`输出图片 ${index + 1}`}
                                loading="lazy"
                                sx={{
                                  width: 96,
                                  height: 96,
                                  objectFit: 'cover',
                                  borderRadius: 1,
                                  border: '1px solid',
                                  borderColor: 'divider',
                                  bgcolor: 'background.paper',
                                }}
                              />
                            ))}
                          </Box>
                        </Box>
                      )}

                      <Divider sx={{ my: 1 }} />

                      <Typography variant="body2" color="text.secondary">
                        记录 ID：{record.id}
                      </Typography>
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
      </Container>
    </Box>
  );
};

export default GenerationHistoryPage;
