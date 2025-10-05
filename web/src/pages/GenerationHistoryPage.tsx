import React from 'react';
import { Box, Container, Typography } from '@mui/material';

const GenerationHistoryPage: React.FC = () => (
  <Box sx={{ py: 8 }}>
    <Container maxWidth="md" sx={{ textAlign: 'center' }}>
      <Typography variant="h4" gutterBottom>
        生成记录
      </Typography>
      <Typography variant="body1" color="text.secondary">
        功能开发中，敬请期待。
      </Typography>
    </Container>
  </Box>
);

export default GenerationHistoryPage;
