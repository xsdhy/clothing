import React, { useCallback, useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Link,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import {
  NavLink,
  useLocation,
  useNavigate,
  type Location,
} from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, ensureStatus } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let active = true;
    void (async () => {
      try {
        const hasUser = await ensureStatus();
        if (!hasUser && active) {
          navigate("/auth/register", { replace: true });
        }
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : "检查系统状态失败");
        }
      }
    })();
    return () => {
      active = false;
    };
  }, [ensureStatus, navigate]);

  const handleSubmit = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault();
      if (!email.trim() || !password.trim()) {
        setError("请输入邮箱和密码");
        return;
      }
      setSubmitting(true);
      setError(null);
      try {
        await login(email.trim(), password);
        const redirectTo =
          (location.state as { from?: Location })?.from?.pathname ?? "/custom";
        navigate(redirectTo, { replace: true });
      } catch (err) {
        setError(err instanceof Error ? err.message : "登录失败，请重试");
      } finally {
        setSubmitting(false);
      }
    },
    [email, password, login, navigate, location.state],
  );

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background:
          "linear-gradient(180deg, rgba(248,249,250,1) 0%, rgba(248,249,250,0.92) 60%, rgba(255,255,255,1) 100%)",
        px: 2,
      }}
    >
      <Paper
        component="form"
        onSubmit={handleSubmit}
        elevation={8}
        sx={{
          width: "100%",
          maxWidth: 420,
          p: 4,
          display: "flex",
          flexDirection: "column",
          gap: 3,
        }}
      >
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
            欢迎回来
          </Typography>
          <Typography variant="body1" color="text.secondary">
            登录以继续使用 AI 图片工作台
          </Typography>
        </Box>

        {error && <Alert severity="error">{error}</Alert>}

        <Stack spacing={2.5}>
          <TextField
            label="邮箱"
            type="email"
            fullWidth
            value={email}
            autoComplete="email"
            onChange={(event) => setEmail(event.target.value)}
            required
          />
          <TextField
            label="密码"
            type="password"
            autoComplete="current-password"
            fullWidth
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            required
          />
        </Stack>

        <Button
          type="submit"
          variant="contained"
          size="large"
          disabled={submitting}
          sx={{ mt: 1 }}
        >
          {submitting ? "登录中..." : "登录"}
        </Button>

        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ textAlign: "center" }}
        >
          首次使用？
          <Link component={NavLink} to="/auth/register" sx={{ ml: 1 }}>
            初始化管理员
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
};

export default LoginPage;
