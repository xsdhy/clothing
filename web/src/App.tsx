import React, { useState } from 'react';
import {
  AppBar,
  Box,
  Button,
  Container,
  CssBaseline,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  ThemeProvider,
  Toolbar,
  Typography,
  createTheme,
} from '@mui/material';
import { NavLink, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import MenuRoundedIcon from '@mui/icons-material/MenuRounded';
import CloseRoundedIcon from '@mui/icons-material/CloseRounded';
import SettingsRoundedIcon from '@mui/icons-material/SettingsRounded';

import AdvancedSettingsPage from './pages/AdvancedSettingsPage';
import CustomImageGenerationPage from './pages/CustomImageGenerationPage';
import GenerationHistoryPage from './pages/GenerationHistoryPage';
import GeneratedImageGalleryPage from './pages/GeneratedImageGalleryPage';
import SceneImageGenerationPage from './pages/SceneImageGenerationPage';

const theme = createTheme({
  typography: {
    fontFamily: '"Noto Sans SC","Inter","Roboto","Helvetica","Arial",sans-serif',
    h4: {
      fontWeight: 700,
    },
    h5: {
      fontWeight: 600,
    },
    button: {
      fontWeight: 600,
    },
  },
  palette: {
    mode: 'light',
    primary: {
      main: '#6366F1',
      light: '#8B90F6',
      dark: '#4F51C9',
    },
    secondary: {
      main: '#F97316',
      light: '#FDBA74',
      dark: '#C2410C',
    },
    background: {
      default: '#F8F9FA',
      paper: '#FFFFFF',
    },
    text: {
      primary: '#1F2937',
      secondary: '#6B7280',
    },
    divider: '#E5E7EB',
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: '#F8F9FA',
        },
      },
    },
    MuiAppBar: {
      defaultProps: {
        elevation: 0,
      },
      styleOverrides: {
        root: {
          backgroundColor: '#FFFFFF',
          color: '#1F2937',
          borderBottom: '1px solid #E5E7EB',
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 12,
          boxShadow: '0 12px 40px rgba(99, 102, 241, 0.08)',
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          textTransform: 'none',
          fontSize: '0.95rem',
          padding: '10px 18px',
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
  { label: '自定义生成', path: '/custom' },
  { label: '场景生成', path: '/scene' },
  { label: '生成记录', path: '/history' },
  { label: '瀑布流', path: '/gallery' },
];

const App: React.FC = () => {
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const isRouteActive = (path: string) => {
    if (path === '/') {
      return location.pathname === path;
    }
    return location.pathname === path || location.pathname.startsWith(`${path}/`);
  };

  const handleToggleMobileMenu = () => {
    setMobileMenuOpen((prev) => !prev);
  };

  const handleCloseMobileMenu = () => {
    setMobileMenuOpen(false);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          background:
            'linear-gradient(180deg, rgba(248,249,250,1) 0%, rgba(248,249,250,0.92) 60%, rgba(255,255,255,1) 100%)',
        }}
      >
        <AppBar position="sticky" color="transparent">
          <Container maxWidth="lg">
            <Toolbar
              disableGutters
              sx={{
                minHeight: 72,
                display: 'flex',
                justifyContent: 'space-between',
                gap: 2,
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                <Box
                  sx={{
                    width: 36,
                    height: 36,
                    borderRadius: '50%',
                    background: 'linear-gradient(135deg, rgba(99,102,241,1) 0%, rgba(129,140,248,1) 100%)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: '#FFFFFF',
                    boxShadow: '0 10px 25px rgba(99,102,241,0.35)',
                  }}
                >
                  <AutoAwesomeIcon sx={{ fontSize: 22 }} />
                </Box>
                <Box>
                  <Typography variant="h6" component="div" sx={{ fontWeight: 700 }}>
                    AI 图片工作台
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    用创意点亮你的灵感瞬间
                  </Typography>
                </Box>
              </Box>

              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                <Box
                  sx={{
                    display: { xs: 'none', md: 'flex' },
                    gap: 1,
                    alignItems: 'center',
                  }}
                >
                  {navItems.map((item) => {
                    const active = isRouteActive(item.path);
                    return (
                      <Button
                        key={item.path}
                        component={NavLink}
                        to={item.path}
                        disableRipple
                        sx={{
                          color: active ? 'primary.main' : 'text.secondary',
                          backgroundColor: active ? 'rgba(99,102,241,0.08)' : 'transparent',
                          boxShadow: active ? '0 8px 20px rgba(99,102,241,0.15)' : 'none',
                          fontWeight: 600,
                          '&:hover': {
                            color: 'primary.main',
                            backgroundColor: 'rgba(99,102,241,0.08)',
                          },
                        }}
                      >
                        {item.label}
                      </Button>
                    );
                  })}
                </Box>

                <IconButton
                  component={NavLink}
                  to="/settings"
                  disableRipple
                  sx={{
                    display: { xs: 'none', md: 'inline-flex' },
                    color: isRouteActive('/settings') ? 'primary.main' : 'text.secondary',
                    backgroundColor: isRouteActive('/settings')
                      ? 'rgba(99,102,241,0.16)'
                      : 'rgba(99,102,241,0.08)',
                    boxShadow: isRouteActive('/settings')
                      ? '0 12px 26px rgba(99,102,241,0.22)'
                      : '0 10px 24px rgba(99,102,241,0.18)',
                    '&:hover': {
                      color: 'primary.main',
                      backgroundColor: 'rgba(99,102,241,0.16)',
                    },
                  }}
                  aria-label="高级设置"
                >
                  <SettingsRoundedIcon />
                </IconButton>

                <IconButton
                  onClick={handleToggleMobileMenu}
                  sx={{
                    display: { xs: 'inline-flex', md: 'none' },
                    color: 'text.secondary',
                    backgroundColor: 'rgba(99,102,241,0.08)',
                    boxShadow: '0 10px 24px rgba(99,102,241,0.18)',
                    '&:hover': {
                      color: 'primary.main',
                      backgroundColor: 'rgba(99,102,241,0.16)',
                    },
                  }}
                  aria-label="打开导航菜单"
                >
                  <MenuRoundedIcon />
                </IconButton>
              </Box>
            </Toolbar>
          </Container>
        </AppBar>

        <Drawer
          anchor="right"
          open={mobileMenuOpen}
          onClose={handleCloseMobileMenu}
          PaperProps={{
            sx: {
              width: '80vw',
              maxWidth: 320,
              background:
                'linear-gradient(180deg, rgba(255,255,255,1) 0%, rgba(244,246,255,0.95) 100%)',
              borderLeft: '1px solid',
              borderColor: 'divider',
              boxShadow: '0 20px 60px rgba(15,23,42,0.18)',
              p: 3,
              display: 'flex',
              flexDirection: 'column',
              gap: 2,
            },
          }}
        >
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <Box>
              <Typography variant="subtitle2" color="text.secondary">
                导航
              </Typography>
              <Typography variant="h6" sx={{ fontWeight: 700 }}>
                AI 图片工作台
              </Typography>
            </Box>
            <IconButton
              onClick={handleCloseMobileMenu}
              sx={{
                color: 'text.secondary',
                backgroundColor: 'rgba(99,102,241,0.08)',
                '&:hover': {
                  color: 'primary.main',
                  backgroundColor: 'rgba(99,102,241,0.16)',
                },
              }}
              aria-label="关闭导航菜单"
            >
              <CloseRoundedIcon />
            </IconButton>
          </Box>

          <List sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
            {navItems.map((item) => {
              const active = isRouteActive(item.path);
              return (
                <ListItemButton
                  key={item.path}
                  component={NavLink}
                  to={item.path}
                  onClick={handleCloseMobileMenu}
                  disableRipple
                  sx={{
                    borderRadius: 2,
                    px: 2,
                    py: 1.5,
                    backgroundColor: active ? 'rgba(99,102,241,0.12)' : 'transparent',
                    boxShadow: active ? '0 16px 32px rgba(99,102,241,0.18)' : 'none',
                    color: active ? 'primary.main' : 'text.primary',
                    '&:hover': {
                      backgroundColor: 'rgba(99,102,241,0.12)',
                      color: 'primary.main',
                    },
                  }}
                >
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{
                      fontWeight: 600,
                    }}
                  />
                </ListItemButton>
              );
            })}
          </List>

          <Divider sx={{ my: 1 }} />

          <ListItemButton
            component={NavLink}
            to="/settings"
            onClick={handleCloseMobileMenu}
            disableRipple
            sx={{
              borderRadius: 2,
              px: 2,
              py: 1.5,
              alignSelf: 'flex-start',
              backgroundColor: isRouteActive('/settings') ? 'rgba(99,102,241,0.12)' : 'transparent',
              boxShadow: isRouteActive('/settings') ? '0 16px 32px rgba(99,102,241,0.18)' : 'none',
              color: isRouteActive('/settings') ? 'primary.main' : 'text.primary',
              '&:hover': {
                backgroundColor: 'rgba(99,102,241,0.12)',
                color: 'primary.main',
              },
            }}
          >
            <ListItemText
              primary="高级设置"
              primaryTypographyProps={{
                fontWeight: 600,
              }}
            />
          </ListItemButton>
        </Drawer>

        <Box component="main" sx={{ flex: 1, py: { xs: 4, md: 6 } }}>
          <Container maxWidth="lg">
            <Routes>
              <Route path="/" element={<Navigate to="/custom" replace />} />
              <Route path="/custom" element={<CustomImageGenerationPage />} />
              <Route path="/scene" element={<SceneImageGenerationPage />} />
              <Route path="/history" element={<GenerationHistoryPage />} />
              <Route path="/gallery" element={<GeneratedImageGalleryPage />} />
              <Route path="/settings" element={<AdvancedSettingsPage />} />
              <Route path="*" element={<Navigate to="/custom" replace />} />
            </Routes>
          </Container>
        </Box>
      </Box>
    </ThemeProvider>
  );
};

export default App;
