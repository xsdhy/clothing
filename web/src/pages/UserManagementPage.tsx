import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
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
import type { UpdateUserPayload } from "../api/auth";
import { createUser, deleteUser, listUsers, updateUser } from "../api/auth";

const roleLabels: Record<string, string> = {
  super_admin: "超级管理员",
  admin: "管理员",
  user: "普通用户",
};

interface CreateUserFormState {
  email: string;
  password: string;
  displayName: string;
  role: string;
  isActive: boolean;
}

interface EditUserFormState {
  displayName: string;
  role: string;
  isActive: boolean;
}

const UserManagementPage: React.FC = () => {
  const { isAdmin, isSuperAdmin, user: currentUser } = useAuth();
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [createDialogError, setCreateDialogError] = useState<string | null>(
    null,
  );
  const [creating, setCreating] = useState(false);
  const [createForm, setCreateForm] = useState<CreateUserFormState>(() => ({
    email: "",
    password: "",
    displayName: "",
    role: isSuperAdmin ? "admin" : "user",
    isActive: true,
  }));

  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editDialogError, setEditDialogError] = useState<string | null>(null);
  const [editing, setEditing] = useState(false);
  const [editingUser, setEditingUser] = useState<UserSummary | null>(null);
  const [editForm, setEditForm] = useState<EditUserFormState>({
    displayName: "",
    role: isSuperAdmin ? "admin" : "user",
    isActive: true,
  });

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

  const buildCreateForm = useCallback(
    (): CreateUserFormState => ({
      email: "",
      password: "",
      displayName: "",
      role: isSuperAdmin ? "admin" : "user",
      isActive: true,
    }),
    [isSuperAdmin],
  );

  const openCreateDialog = useCallback(() => {
    setCreateForm(buildCreateForm());
    setCreateDialogError(null);
    setCreateDialogOpen(true);
  }, [buildCreateForm]);

  const closeCreateDialog = useCallback(() => {
    if (creating) {
      return;
    }
    setCreateDialogOpen(false);
    setCreateDialogError(null);
  }, [creating]);

  const handleSubmitCreate = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      const trimmedEmail = createForm.email.trim();
      const trimmedPassword = createForm.password.trim();
      const trimmedDisplayName = createForm.displayName.trim();

      if (!trimmedEmail || !trimmedPassword) {
        setCreateDialogError("请输入完整的用户信息");
        return;
      }
      if (trimmedPassword.length < 8) {
        setCreateDialogError("密码至少需要 8 个字符");
        return;
      }

      setCreating(true);
      setCreateDialogError(null);
      try {
        await createUser({
          email: trimmedEmail,
          password: trimmedPassword,
          display_name: trimmedDisplayName || undefined,
          role: createForm.role,
          is_active: createForm.isActive,
        });
        setCreateDialogOpen(false);
        setCreateForm(buildCreateForm());
        await loadUsers();
      } catch (err) {
        setCreateDialogError(
          err instanceof Error ? err.message : "创建用户失败",
        );
      } finally {
        setCreating(false);
      }
    },
    [buildCreateForm, createForm, loadUsers],
  );

  const canEditUser = useCallback(
    (user: UserSummary): boolean =>
      user.role !== "super_admin" && (isSuperAdmin || user.role === "user"),
    [isSuperAdmin],
  );

  const openEditDialog = useCallback(
    (user: UserSummary) => {
      if (!canEditUser(user)) {
        return;
      }
      setEditingUser(user);
      setEditForm({
        displayName: user.display_name ?? "",
        role: user.role,
        isActive: user.is_active,
      });
      setEditDialogError(null);
      setEditDialogOpen(true);
    },
    [canEditUser],
  );

  const closeEditDialog = useCallback(() => {
    if (editing) {
      return;
    }
    setEditDialogOpen(false);
    setEditingUser(null);
    setEditDialogError(null);
  }, [editing]);

  const editRoleOptions = useMemo(() => {
    if (!editingUser) {
      return availableRoles;
    }
    const options = [...availableRoles];
    if (!options.some((role) => role.value === editingUser.role)) {
      options.push({
        value: editingUser.role,
        label: roleLabels[editingUser.role] ?? editingUser.role,
      });
    }
    return options;
  }, [availableRoles, editingUser]);

  const handleSubmitEdit = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      if (!editingUser) {
        return;
      }
      const trimmedDisplayName = editForm.displayName.trim();
      const payload: UpdateUserPayload = {};

      if (trimmedDisplayName !== (editingUser.display_name ?? "")) {
        payload.display_name = trimmedDisplayName || null;
      }
      if (isSuperAdmin && editForm.role !== editingUser.role) {
        payload.role = editForm.role;
      }
      if (editForm.isActive !== editingUser.is_active) {
        payload.is_active = editForm.isActive;
      }

      if (Object.keys(payload).length === 0) {
        setEditDialogError("未检测到改动");
        return;
      }

      setEditing(true);
      setEditDialogError(null);
      try {
        await updateUser(editingUser.id, payload);
        setEditDialogOpen(false);
        setEditingUser(null);
        await loadUsers();
      } catch (err) {
        setEditDialogError(
          err instanceof Error ? err.message : "更新用户失败",
        );
      } finally {
        setEditing(false);
      }
    },
    [editForm.isActive, editForm.role, editForm.displayName, editingUser, isSuperAdmin, loadUsers],
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
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <Box>
        <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.5 }}>
          用户管理
        </Typography>
        <Typography variant="body2" color="text.secondary">
          超级管理员可创建、停用或删除用户，管理员可管理普通用户账号。
        </Typography>
      </Box>
      <Paper sx={{ p: 3 }}>
        <Stack
          direction={{ xs: "column", sm: "row" }}
          spacing={2}
          justifyContent="space-between"
          alignItems={{ xs: "stretch", sm: "center" }}
          sx={{ mb: 2 }}
        >
          <Box>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              用户列表
            </Typography>
            <Typography variant="body2" color="text.secondary">
              支持快速停用、重置密码或删除账号。
            </Typography>
          </Box>
          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              onClick={() => void loadUsers()}
              disabled={loading}
            >
              刷新
            </Button>
            <Button
              variant="contained"
              onClick={openCreateDialog}
              disabled={creating}
            >
              新增用户
            </Button>
          </Stack>
        </Stack>

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
          <Table size="small">
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
                const canEdit = canEditUser(user);
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
                        {canEdit && (
                          <Button
                            variant="outlined"
                            size="small"
                            onClick={() => openEditDialog(user)}
                            disabled={editing}
                          >
                            编辑
                          </Button>
                        )}
                        {!isSuper && (
                          <Button
                            variant="text"
                            size="small"
                            onClick={() => void handleToggleActive(user)}
                            disabled={editing}
                          >
                            {user.is_active ? "停用" : "启用"}
                          </Button>
                        )}
                        {!isSuper && (
                          <Button
                            variant="text"
                            size="small"
                            onClick={() => void handleResetPassword(user)}
                            disabled={editing}
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
                            disabled={editing}
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

      <Dialog
        open={createDialogOpen}
        onClose={closeCreateDialog}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle>新增用户</DialogTitle>
        <Box component="form" onSubmit={handleSubmitCreate}>
          <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {createDialogError && <Alert severity="error">{createDialogError}</Alert>}
            <TextField
              label="邮箱"
              type="email"
              value={createForm.email}
              onChange={(event) =>
                setCreateForm((prev) => ({ ...prev, email: event.target.value }))
              }
              required
            />
            <TextField
              label="临时密码"
              type="password"
              value={createForm.password}
              onChange={(event) =>
                setCreateForm((prev) => ({
                  ...prev,
                  password: event.target.value,
                }))
              }
              helperText="至少 8 个字符"
              required
            />
            <TextField
              label="显示昵称（可选）"
              value={createForm.displayName}
              onChange={(event) =>
                setCreateForm((prev) => ({
                  ...prev,
                  displayName: event.target.value,
                }))
              }
            />
            <FormControl fullWidth>
              <InputLabel id="create-user-role-label">角色</InputLabel>
              <Select
                labelId="create-user-role-label"
                label="角色"
                value={createForm.role}
                onChange={(event: SelectChangeEvent<string>) =>
                  setCreateForm((prev) => ({
                    ...prev,
                    role: event.target.value,
                  }))
                }
              >
                {availableRoles.map((role) => (
                  <MenuItem key={role.value} value={role.value}>
                    {role.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControlLabel
              control={
                <Switch
                  checked={createForm.isActive}
                  onChange={(event) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      isActive: event.target.checked,
                    }))
                  }
                />
              }
              label={createForm.isActive ? "启用状态" : "停用状态"}
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3 }}>
            <Button onClick={closeCreateDialog} disabled={creating}>
              取消
            </Button>
            <Button type="submit" variant="contained" disabled={creating}>
              {creating ? "创建中..." : "创建"}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      <Dialog
        open={editDialogOpen}
        onClose={closeEditDialog}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle>
          编辑用户{editingUser ? `：${editingUser.email}` : ""}
        </DialogTitle>
        <Box component="form" onSubmit={handleSubmitEdit}>
          <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {editDialogError && <Alert severity="error">{editDialogError}</Alert>}
            <TextField
              label="邮箱"
              value={editingUser?.email ?? ""}
              disabled
            />
            <TextField
              label="显示昵称（可选）"
              value={editForm.displayName}
              onChange={(event) =>
                setEditForm((prev) => ({
                  ...prev,
                  displayName: event.target.value,
                }))
              }
            />
            <FormControl fullWidth disabled={!isSuperAdmin}>
              <InputLabel id="edit-user-role-label">角色</InputLabel>
              <Select
                labelId="edit-user-role-label"
                label="角色"
                value={editForm.role}
                onChange={(event: SelectChangeEvent<string>) =>
                  setEditForm((prev) => ({
                    ...prev,
                    role: event.target.value,
                  }))
                }
              >
                {editRoleOptions.map((role) => (
                  <MenuItem key={role.value} value={role.value}>
                    {role.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControlLabel
              control={
                <Switch
                  checked={editForm.isActive}
                  onChange={(event) =>
                    setEditForm((prev) => ({
                      ...prev,
                      isActive: event.target.checked,
                    }))
                  }
                />
              }
              label={editForm.isActive ? "启用状态" : "停用状态"}
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3 }}>
            <Button onClick={closeEditDialog} disabled={editing}>
              取消
            </Button>
            <Button type="submit" variant="contained" disabled={editing}>
              {editing ? "保存中..." : "保存"}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>
    </Box>
  );
};

export default UserManagementPage;
