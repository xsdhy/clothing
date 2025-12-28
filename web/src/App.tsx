import React, { useMemo, useState } from "react";
import {
  AppBar,
  Avatar,
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
  CircularProgress,
  createTheme,
} from "@mui/material";
import {
  NavLink,
  Navigate,
  Outlet,
  Route,
  Routes,
  useLocation,
} from "react-router-dom";
import AutoAwesomeIcon from "@mui/icons-material/AutoAwesome";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";

import CustomImageGenerationPage from "./pages/CustomImageGenerationPage";
import SceneImageGenerationPage from "./pages/SceneImageGenerationPage";
import GenerationHistoryPage from "./pages/GenerationHistoryPage";
import GeneratedImageGalleryPage from "./pages/GeneratedImageGalleryPage";
import AdvancedSettingsPage from "./pages/AdvancedSettingsPage";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import UserManagementPage from "./pages/UserManagementPage";
import ProviderManagementPage from "./pages/ProviderManagementPage";
import TagManagementPage from "./pages/TagManagementPage";
import { useAuth } from "./contexts/AuthContext";

const theme = createTheme({
  typography: {
    fontFamily:
      '"Noto Sans SC","Inter","Roboto","Helvetica","Arial",sans-serif',
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
    mode: "light",
    primary: {
      main: "#6366F1",
      light: "#8B90F6",
      dark: "#4F51C9",
    },
    secondary: {
      main: "#F97316",
      light: "#FDBA74",
      dark: "#C2410C",
    },
    background: {
      default: "#F8F9FA",
      paper: "#FFFFFF",
    },
    text: {
      primary: "#1F2937",
      secondary: "#6B7280",
    },
    divider: "#E5E7EB",
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: "#F8F9FA",
        },
      },
    },
    MuiAppBar: {
      defaultProps: {
        elevation: 0,
      },
      styleOverrides: {
        root: {
          backgroundColor: "#FFFFFF",
          color: "#1F2937",
          borderBottom: "1px solid #E5E7EB",
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          boxShadow: "0 4px 20px rgba(99, 102, 241, 0.05)",
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          textTransform: "none",
          fontSize: "0.9rem",
          padding: "6px 16px",
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 600,
          fontSize: "0.8125rem",
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          padding: "12px 16px",
        },
      },
    },
  },
});

const FullscreenLoader: React.FC = () => (
  <Box
    sx={{
      minHeight: "100vh",
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      backgroundColor: "background.default",
    }}
  >
    <CircularProgress color="primary" />
  </Box>
);

const RequireAuth: React.FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  const { loading, isAuthenticated } = useAuth();
  const location = useLocation();

  if (loading) {
    return <FullscreenLoader />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" replace state={{ from: location }} />;
  }

  return children;
};

const GuestOnly: React.FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  const { loading, isAuthenticated } = useAuth();

  if (loading) {
    return <FullscreenLoader />;
  }

  if (isAuthenticated) {
    return <Navigate to="/custom" replace />;
  }

  return children;
};

const RequireAdmin: React.FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  const { loading, isAuthenticated, isAdmin } = useAuth();
  const location = useLocation();

  if (loading) {
    return <FullscreenLoader />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" replace state={{ from: location }} />;
  }

  if (!isAdmin) {
    return <Navigate to="/settings" replace />;
  }

  return children;
};

const AppLayout: React.FC = () => {
  const location = useLocation();
  const { user, logout, isAdmin } = useAuth();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navItems = useMemo(() => {
    const items = [
      { label: "自定义生成", path: "/custom" },
      { label: "场景生成", path: "/scene" },
      { label: "生成记录", path: "/history" },
      { label: "瀑布流", path: "/gallery" },
    ];
    if (isAdmin) {
      items.push({ label: "厂商管理", path: "/providers" });
      items.push({ label: "标签管理", path: "/tags" });
      items.push({ label: "用户管理", path: "/users" });
    }
    return items;
  }, [isAdmin]);

  const isRouteActive = (path: string) => {
    if (path === "/") {
      return location.pathname === path;
    }
    return (
      location.pathname === path || location.pathname.startsWith(`${path}/`)
    );
  };

  const handleToggleMobileMenu = () => {
    setMobileMenuOpen((prev) => !prev);
  };

  const handleCloseMobileMenu = () => {
    setMobileMenuOpen(false);
  };

  const handleLogout = () => {
    logout();
    handleCloseMobileMenu();
  };

  const displayName = user?.display_name || user?.email || "用户";

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        background:
          "linear-gradient(180deg, rgba(248,249,250,1) 0%, rgba(248,249,250,0.92) 80%, rgba(255,255,255,1) 100%)",
      }}
    >
      <AppBar position="sticky" color="transparent">
        <Container maxWidth="lg">
          <Toolbar
            disableGutters
            sx={{
              minHeight: 64,
              display: "flex",
              justifyContent: "space-between",
              gap: 2,
            }}
          >
            <Box sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
              <Box
                sx={{
                  width: 32,
                  height: 32,
                  borderRadius: "50%",
                  background:
                    "linear-gradient(135deg, rgba(99,102,241,1) 0%, rgba(129,140,248,1) 100%)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  color: "#FFFFFF",
                  boxShadow: "0 8px 20px rgba(99,102,241,0.25)",
                }}
              >
                <AutoAwesomeIcon sx={{ fontSize: 18 }} />
              </Box>
              <Box>
                <Typography
                  variant="subtitle1"
                  component="div"
                  sx={{ fontWeight: 700, lineHeight: 1.2 }}
                >
                  AI 图片工作台
                </Typography>
              </Box>
            </Box>

            <Box sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
              <Box
                sx={{
                  display: { xs: "none", md: "flex" },
                  gap: 0.5,
                  alignItems: "center",
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
                      size="small"
                      sx={{
                        color: active ? "primary.main" : "text.secondary",
                        backgroundColor: active
                          ? "rgba(99,102,241,0.08)"
                          : "transparent",
                        boxShadow: "none",
                        fontWeight: 600,
                        "&:hover": {
                          color: "primary.main",
                          backgroundColor: "rgba(99,102,241,0.08)",
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
                size="small"
                sx={{
                  display: { xs: "none", md: "inline-flex" },
                  color: isRouteActive("/settings")
                    ? "primary.main"
                    : "text.secondary",
                  backgroundColor: isRouteActive("/settings")
                    ? "rgba(99,102,241,0.12)"
                    : "transparent",
                  "&:hover": {
                    color: "primary.main",
                    backgroundColor: "rgba(99,102,241,0.12)",
                  },
                }}
                aria-label="高级设置"
              >
                <SettingsRoundedIcon fontSize="small" />
              </IconButton>

              <Box
                sx={{
                  display: { xs: "none", md: "flex" },
                  alignItems: "center",
                  gap: 1,
                  px: 1,
                  py: 0.5,
                  borderRadius: 2,
                  backgroundColor: "rgba(99,102,241,0.06)",
                }}
              >
                <Avatar sx={{ width: 28, height: 28, bgcolor: "primary.main", fontSize: "0.8rem" }}>
                  {displayName?.charAt(0)?.toUpperCase() ?? "?"}
                </Avatar>
                <Box sx={{ display: "flex", flexDirection: "column" }}>
                  <Typography variant="caption" sx={{ fontWeight: 600 }}>
                    {displayName}
                  </Typography>
                </Box>
                <Button
                  size="small"
                  variant="text"
                  sx={{
                    minWidth: 0,
                    p: 0,
                    ml: 0.5,
                    color: "text.secondary",
                    fontSize: "0.75rem"
                  }}
                  onClick={logout}
                >
                  退出
                </Button>
              </Box>

              <IconButton
                onClick={handleToggleMobileMenu}
                sx={{
                  display: { xs: "inline-flex", md: "none" },
                  color: "text.secondary",
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
            width: "75vw",
            maxWidth: 280,
            background: "#FFFFFF",
            p: 2,
            display: "flex",
            flexDirection: "column",
            gap: 2,
          },
        }}
      >
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
            AI 图片工作台
          </Typography>
          <IconButton
            onClick={handleCloseMobileMenu}
            size="small"
            aria-label="关闭导航菜单"
          >
            <CloseRoundedIcon />
          </IconButton>
        </Box>

        <List sx={{ display: "flex", flexDirection: "column", gap: 0.5 }}>
          {navItems.map((item) => {
            const active = isRouteActive(item.path);
            return (
              <ListItemButton
                key={item.path}
                component={NavLink}
                to={item.path}
                onClick={handleCloseMobileMenu}
                disableRipple
                selected={active}
                sx={{
                  borderRadius: 2,
                  py: 1,
                  "&.Mui-selected": {
                    backgroundColor: "rgba(99,102,241,0.12)",
                    color: "primary.main",
                    "&:hover": {
                      backgroundColor: "rgba(99,102,241,0.16)",
                    }
                  },
                }}
              >
                <ListItemText
                  primary={item.label}
                  primaryTypographyProps={{
                    fontWeight: active ? 600 : 400,
                    fontSize: "0.9rem",
                  }}
                />
              </ListItemButton>
            );
          })}
        </List>

        <ListItemButton
          component={NavLink}
          to="/settings"
          onClick={handleCloseMobileMenu}
          disableRipple
          selected={isRouteActive("/settings")}
          sx={{
            borderRadius: 2,
            py: 1,
            "&.Mui-selected": {
              backgroundColor: "rgba(99,102,241,0.12)",
              color: "primary.main",
            }
          }}
        >
          <ListItemText
            primary="高级设置"
            primaryTypographyProps={{
              fontWeight: 600,
              fontSize: "0.9rem",
            }}
          />
        </ListItemButton>

        <Divider sx={{ my: 1 }} />

        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            gap: 1.5,
            backgroundColor: "rgba(99,102,241,0.06)",
            borderRadius: 2,
            px: 2,
            py: 1.5,
          }}
        >
          <Avatar sx={{ width: 32, height: 32, bgcolor: "primary.main", fontSize: "0.9rem" }}>
            {displayName?.charAt(0)?.toUpperCase() ?? "?"}
          </Avatar>
          <Box sx={{ display: "flex", flexDirection: "column", flex: 1 }}>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {displayName}
            </Typography>
            <Button
              size="small"
              variant="text"
              sx={{ alignSelf: "flex-start", p: 0, minWidth: 0, color: "error.main", fontSize: "0.75rem" }}
              onClick={handleLogout}
            >
              退出登录
            </Button>
          </Box>
        </Box>
      </Drawer>

      <Box component="main" sx={{ flex: 1, py: { xs: 2.5, md: 3.5 } }}>
        <Container maxWidth="lg">
          <Outlet />
        </Container>
      </Box>
    </Box>
  );
};

const App: React.FC = () => (
  <ThemeProvider theme={theme}>
    <CssBaseline />
    <Routes>
      <Route
        path="/auth/login"
        element={
          <GuestOnly>
            <LoginPage />
          </GuestOnly>
        }
      />
      <Route
        path="/auth/register"
        element={
          <GuestOnly>
            <RegisterPage />
          </GuestOnly>
        }
      />
      <Route
        path="/"
        element={
          <RequireAuth>
            <AppLayout />
          </RequireAuth>
        }
      >
        <Route index element={<Navigate to="custom" replace />} />
        <Route path="custom" element={<CustomImageGenerationPage />} />
        <Route path="scene" element={<SceneImageGenerationPage />} />
        <Route path="history" element={<GenerationHistoryPage />} />
        <Route path="gallery" element={<GeneratedImageGalleryPage />} />
        <Route path="settings" element={<AdvancedSettingsPage />} />
        <Route
          path="providers"
          element={
            <RequireAdmin>
              <ProviderManagementPage />
            </RequireAdmin>
          }
        />
        <Route
          path="tags"
          element={
            <RequireAdmin>
              <TagManagementPage />
            </RequireAdmin>
          }
        />
        <Route
          path="users"
          element={
            <RequireAdmin>
              <UserManagementPage />
            </RequireAdmin>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  </ThemeProvider>
);

export default App;
