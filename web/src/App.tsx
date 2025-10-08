import React from 'react';
import {
  AppBar,
  Box,
  Button,
  CssBaseline,
  ThemeProvider,
  Toolbar,
  Typography,
  createTheme,
} from '@mui/material';
import { NavLink, Navigate, Route, Routes } from 'react-router-dom';

import AdvancedSettingsPage from './pages/AdvancedSettingsPage';
import CustomImageGenerationPage from './pages/CustomImageGenerationPage';
import GenerationHistoryPage from './pages/GenerationHistoryPage';
import GeneratedImageGalleryPage from './pages/GeneratedImageGalleryPage';
import SceneImageGenerationPage from './pages/SceneImageGenerationPage';

const theme = createTheme({
  palette: {
    primary: {
      main: '#3f51b5',
      light: '#7986cb',
      dark: '#303f9f',
    },
    secondary: {
      main: '#f50057',
      light: '#ff5983',
      dark: '#c51162',
    },
    background: {
      default: '#f5f5f5',
      paper: '#ffffff',
    },
  },
  typography: {
    h4: {
      fontWeight: 600,
    },
    h5: {
      fontWeight: 600,
    },
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 8,
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 6,
          textTransform: 'none',
          fontWeight: 600,
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 600,
        },
      },
    },
  },
});

const navItems = [
  { label: '自定义图片生成', path: '/custom' },
  { label: '场景图片生成', path: '/scene' },
  { label: '生成记录查看', path: '/history' },
  { label: '生成图片查看', path: '/gallery' },
  { label: '高级设置', path: '/settings' },
];

const App: React.FC = () => (
  <ThemeProvider theme={theme}>
    <CssBaseline />
    <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <AppBar position="static" color="primary">
        <Toolbar sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, justifyContent: 'space-between' }}>
          <Typography variant="h6" component="div" sx={{ fontWeight: 700 }}>
            AI 图片工作台
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
            {navItems.map((item) => (
              <Button
                key={item.path}
                component={NavLink}
                to={item.path}
                color="inherit"
                sx={{
                  opacity: 0.75,
                  '&.active': {
                    opacity: 1,
                    fontWeight: 700,
                    borderBottom: '2px solid rgba(255,255,255,0.9)',
                    borderRadius: 0,
                  },
                }}
              >
                {item.label}
              </Button>
            ))}
          </Box>
        </Toolbar>
      </AppBar>

      <Box component="main" sx={{ flex: 1, bgcolor: 'background.default' }}>
        <Routes>
          <Route path="/" element={<Navigate to="/custom" replace />} />
          <Route path="/custom" element={<CustomImageGenerationPage />} />
          <Route path="/scene" element={<SceneImageGenerationPage />} />
          <Route path="/history" element={<GenerationHistoryPage />} />
          <Route path="/gallery" element={<GeneratedImageGalleryPage />} />
          <Route path="/settings" element={<AdvancedSettingsPage />} />
          <Route path="*" element={<Navigate to="/custom" replace />} />
        </Routes>
      </Box>
    </Box>
  </ThemeProvider>
);

export default App;
