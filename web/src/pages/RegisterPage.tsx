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
import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const { ensureStatus, registerInitial } = useAuth();

  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let active = true;
    void (async () => {
      try {
        const hasUser = await ensureStatus();
        if (hasUser && active) {
          navigate("/auth/login", { replace: true });
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
      const trimmedEmail = email.trim();
      const trimmedPassword = password.trim();
      if (!trimmedEmail || !trimmedPassword) {
        setError("请输入邮箱和密码");
        return;
      }
      if (trimmedPassword.length < 8) {
        setError("密码至少需要 8 个字符");
        return;
      }
      if (trimmedPassword !== confirmPassword.trim()) {
        setError("两次输入的密码不一致");
        return;
      }

      setSubmitting(true);
      setError(null);
      try {
        await registerInitial({
          email: trimmedEmail,
          password: trimmedPassword,
          displayName: displayName.trim(),
        });
        navigate("/custom", { replace: true });
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "初始化管理员失败，请重试",
        );
      } finally {
        setSubmitting(false);
      }
    },
    [email, password, confirmPassword, displayName, navigate, registerInitial],
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
          maxWidth: 480,
          p: 4,
          display: "flex",
          flexDirection: "column",
          gap: 3,
        }}
      >
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
            初始化超级管理员
          </Typography>
          <Typography variant="body1" color="text.secondary">
            首次使用，请设置超级管理员账号以管理平台
          </Typography>
        </Box>

        {error && <Alert severity="error">{error}</Alert>}

        <Stack spacing={2.5}>
          <TextField
            label="邮箱"
            type="email"
            fullWidth
            autoComplete="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            required
          />
          <TextField
            label="昵称（选填）"
            fullWidth
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
          />
          <TextField
            label="密码"
            type="password"
            autoComplete="new-password"
            fullWidth
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            required
            helperText="至少 8 个字符，建议包含字母与数字"
          />
          <TextField
            label="确认密码"
            type="password"
            autoComplete="new-password"
            fullWidth
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
            required
          />
        </Stack>

        <Button
          type="submit"
          variant="contained"
          size="large"
          disabled={submitting}
        >
          {submitting ? "创建中..." : "创建超级管理员"}
        </Button>

        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ textAlign: "center" }}
        >
          已有账号？
          <Link component={NavLink} to="/auth/login" sx={{ ml: 1 }}>
            前往登录
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
};

export default RegisterPage;
