import React from "react";
import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

const AdvancedSettingsPage: React.FC = () => {
  const navigate = useNavigate();
  const { isAdmin } = useAuth();

  return (
    <Box sx={{ py: 8 }}>
      <Paper sx={{ p: { xs: 4, md: 6 }, maxWidth: 720, margin: "0 auto" }}>
        <Stack spacing={3} alignItems="flex-start">
          <Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
              高级设置
            </Typography>
            <Typography variant="body1" color="text.secondary">
              可在此处集中管理平台的高级功能。配置选项正在规划中，敬请期待。
            </Typography>
          </Box>

          {isAdmin && (
            <Box>
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                管理员工具
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                用户管理已迁移至独立页面，点击下方按钮快速进入。
              </Typography>
              <Button variant="contained" onClick={() => navigate("/users")}>
                前往用户管理
              </Button>
            </Box>
          )}
        </Stack>
      </Paper>
    </Box>
  );
};

export default AdvancedSettingsPage;
