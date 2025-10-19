import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import { useAuth } from "../contexts/AuthContext";
import type { UserSummary } from "../types";
import { createUser, deleteUser, listUsers, updateUser } from "../api/auth";

const roleLabels: Record<string, string> = {
  super_admin: "超级管理员",
  admin: "管理员",
  user: "普通用户",
};

const UserManagementPage: React.FC = () => {
  const { isAdmin, isSuperAdmin, user: currentUser } = useAuth();
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [newEmail, setNewEmail] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [newDisplayName, setNewDisplayName] = useState("");
  const [newRole, setNewRole] = useState(isSuperAdmin ? "admin" : "user");

  const loadUsers = useCallback(async () => {
    if (!isAdmin) {
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const response = await listUsers(1, 100);
      setUsers(response.users ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载用户失败");
    } finally {
      setLoading(false);
    }
  }, [isAdmin]);

  useEffect(() => {
    void loadUsers();
  }, [loadUsers]);

  const availableRoles = useMemo(() => {
    const roles: Array<{ value: string; label: string }> = [
      { value: "user", label: roleLabels.user },
    ];
    if (isSuperAdmin) {
      roles.unshift({ value: "admin", label: roleLabels.admin });
    }
    return roles;
  }, [isSuperAdmin]);

  const handleCreateUser = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault();
      if (!newEmail.trim() || !newPassword.trim()) {
        setCreateError("请输入完整的用户信息");
        return;
      }
      if (newPassword.trim().length < 8) {
        setCreateError("密码至少需要 8 个字符");
        return;
      }
      setCreating(true);
      setCreateError(null);
      try {
        await createUser({
          email: newEmail.trim(),
          password: newPassword.trim(),
          display_name: newDisplayName.trim(),
          role: newRole,
          is_active: true,
        });
        setNewEmail("");
        setNewPassword("");
        setNewDisplayName("");
        setNewRole(isSuperAdmin ? "admin" : "user");
        await loadUsers();
      } catch (err) {
        setCreateError(err instanceof Error ? err.message : "创建用户失败");
      } finally {
        setCreating(false);
      }
    },
    [isSuperAdmin, loadUsers, newDisplayName, newEmail, newPassword, newRole],
  );

  const handleToggleActive = useCallback(
    async (user: UserSummary) => {
      const targetState = !user.is_active;
      try {
        await updateUser(user.id, { is_active: targetState });
        await loadUsers();
      } catch (err) {
        alert(err instanceof Error ? err.message : "更新用户状态失败");
      }
    },
    [loadUsers],
  );

  const handleDelete = useCallback(
    async (user: UserSummary) => {
      if (!window.confirm(`确认删除用户 ${user.email} 吗？`)) {
        return;
      }
      try {
        await deleteUser(user.id);
        await loadUsers();
      } catch (err) {
        alert(err instanceof Error ? err.message : "删除用户失败");
      }
    },
    [loadUsers],
  );

  const handleResetPassword = useCallback(async (user: UserSummary) => {
    const newPwd = window.prompt(`请输入 ${user.email} 的新密码（至少 8 位）`);
    if (!newPwd) {
      return;
    }
    if (newPwd.trim().length < 8) {
      alert("密码至少需要 8 个字符");
      return;
    }
    try {
      await updateUser(user.id, { password: newPwd.trim() });
      alert("密码已更新");
    } catch (err) {
      alert(err instanceof Error ? err.message : "重置密码失败");
    }
  }, []);

  if (!isAdmin) {
    return (
      <Box sx={{ py: 8 }}>
        <Paper sx={{ p: 6, textAlign: "center" }}>
          <Typography variant="h5" sx={{ fontWeight: 600, mb: 2 }}>
            用户管理
          </Typography>
          <Typography variant="body1" color="text.secondary">
            当前账号暂无管理员权限，如需访问，请联系超级管理员。
          </Typography>
        </Paper>
      </Box>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 4 }}>
      <Box>
        <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
          用户管理
        </Typography>
        <Typography variant="body1" color="text.secondary">
          超级管理员可创建、停用或删除用户，管理员可管理普通用户账号。
        </Typography>
      </Box>

      <Paper
        component="form"
        onSubmit={handleCreateUser}
        sx={{ p: 4, display: "flex", flexDirection: "column", gap: 3 }}
      >
        <Box>
          <Typography variant="h6" sx={{ fontWeight: 600, mb: 1 }}>
            创建新用户
          </Typography>
          <Typography variant="body2" color="text.secondary">
            设置登录邮箱、临时密码和角色。密码将在首次登录后提示用户修改。
          </Typography>
        </Box>
        {createError && <Alert severity="error">{createError}</Alert>}
        <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
          <TextField
            label="邮箱"
            type="email"
            required
            fullWidth
            value={newEmail}
            onChange={(event) => setNewEmail(event.target.value)}
          />
          <TextField
            label="临时密码"
            type="password"
            required
            fullWidth
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
            helperText="至少 8 个字符"
          />
        </Stack>
        <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
          <TextField
            label="显示昵称（可选）"
            fullWidth
            value={newDisplayName}
            onChange={(event) => setNewDisplayName(event.target.value)}
          />
          <FormControl fullWidth>
            <InputLabel id="create-user-role-label">角色</InputLabel>
            <Select
              labelId="create-user-role-label"
              label="角色"
              value={newRole}
              onChange={(event: SelectChangeEvent<string>) =>
                setNewRole(event.target.value)
              }
            >
              {availableRoles.map((role) => (
                <MenuItem key={role.value} value={role.value}>
                  {role.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Stack>
        <Button
          type="submit"
          variant="contained"
          size="large"
          disabled={creating}
        >
          {creating ? "创建中..." : "创建用户"}
        </Button>
      </Paper>

      <Paper sx={{ p: 4 }}>
        <Box
          sx={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            mb: 3,
          }}
        >
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            用户列表
          </Typography>
          <Button
            variant="outlined"
            onClick={() => void loadUsers()}
            disabled={loading}
          >
            刷新
          </Button>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {loading ? (
          <Box sx={{ display: "flex", justifyContent: "center", py: 6 }}>
            <CircularProgress />
          </Box>
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>邮箱</TableCell>
                <TableCell>昵称</TableCell>
                <TableCell>角色</TableCell>
                <TableCell>状态</TableCell>
                <TableCell align="right">操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {users.map((user) => {
                const isCurrentUser = currentUser?.id === user.id;
                const isSuper = user.role === "super_admin";
                const isAdminRole = user.role === "admin";
                return (
                  <TableRow key={user.id} hover>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>{user.display_name || "—"}</TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        color={
                          isSuper
                            ? "secondary"
                            : isAdminRole
                              ? "primary"
                              : "default"
                        }
                        label={roleLabels[user.role] ?? user.role}
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        color={user.is_active ? "success" : "default"}
                        label={user.is_active ? "启用" : "停用"}
                      />
                    </TableCell>
                    <TableCell align="right">
                      <Stack
                        direction="row"
                        spacing={1}
                        justifyContent="flex-end"
                      >
                        {!isSuper && (
                          <Button
                            variant="text"
                            size="small"
                            onClick={() => void handleToggleActive(user)}
                          >
                            {user.is_active ? "停用" : "启用"}
                          </Button>
                        )}
                        {!isSuper && (
                          <Button
                            variant="text"
                            size="small"
                            onClick={() => void handleResetPassword(user)}
                          >
                            重置密码
                          </Button>
                        )}
                        {!isSuper && !isCurrentUser && (
                          <Button
                            variant="text"
                            color="error"
                            size="small"
                            onClick={() => void handleDelete(user)}
                          >
                            删除
                          </Button>
                        )}
                        {isSuper && (
                          <Typography variant="body2" color="text.secondary">
                            超级管理员
                          </Typography>
                        )}
                      </Stack>
                    </TableCell>
                  </TableRow>
                );
              })}
              {users.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                    <Typography variant="body2" color="text.secondary">
                      暂无用户数据
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        )}
      </Paper>
    </Box>
  );
};

export default UserManagementPage;
