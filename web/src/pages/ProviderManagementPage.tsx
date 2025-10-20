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
  FormControlLabel,
  Paper,
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
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteForeverRoundedIcon from "@mui/icons-material/DeleteForeverRounded";
import RestartAltRoundedIcon from "@mui/icons-material/RestartAltRounded";
import type {
  ProviderAdmin,
  ProviderModelAdmin,
  ProviderUpdatePayload,
  ProviderModelUpdatePayload,
} from "../types";
import {
  createProvider,
  createProviderModel,
  deleteProvider,
  deleteProviderModel,
  listProvidersAdmin,
  updateProvider,
  updateProviderModel,
} from "../api/providers";
import { useAuth } from "../contexts/AuthContext";

interface ProviderFormState {
  id: string;
  name: string;
  driver: string;
  description: string;
  base_url: string;
  api_key: string;
  config_text: string;
  is_active: boolean;
}

interface ModelDialogForm {
  model_id: string;
  name: string;
  description: string;
  price: string;
  max_images: string;
  modalities: string;
  supported_sizes: string;
  settings: string;
  is_active: boolean;
}

const jsonStringify = (value?: Record<string, unknown>): string =>
  value && Object.keys(value).length > 0 ? JSON.stringify(value, null, 2) : "";

const parseCommaSeparated = (value: string): string[] =>
  value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => item.length > 0);

const defaultProviderForm = (): ProviderFormState => ({
  id: "",
  name: "",
  driver: "",
  description: "",
  base_url: "",
  api_key: "",
  config_text: "",
  is_active: true,
});

const toProviderFormState = (provider: ProviderAdmin): ProviderFormState => ({
  id: provider.id,
  name: provider.name,
  driver: provider.driver,
  description: provider.description ?? "",
  base_url: provider.base_url ?? "",
  api_key: "",
  config_text: jsonStringify(provider.config as Record<string, unknown>),
  is_active: provider.is_active,
});

const defaultModelForm = (): ModelDialogForm => ({
  model_id: "",
  name: "",
  description: "",
  price: "",
  max_images: "",
  modalities: "",
  supported_sizes: "",
  settings: "",
  is_active: true,
});

const toModelDialogForm = (model: ProviderModelAdmin): ModelDialogForm => ({
  model_id: model.model_id,
  name: model.name,
  description: model.description ?? "",
  price: model.price ?? "",
  max_images:
    typeof model.max_images === "number" && Number.isFinite(model.max_images)
      ? String(model.max_images)
      : "",
  modalities: model.modalities?.join(", ") ?? "",
  supported_sizes: model.supported_sizes?.join(", ") ?? "",
  settings: jsonStringify(model.settings as Record<string, unknown>),
  is_active: model.is_active,
});

const ProviderManagementPage: React.FC = () => {
  const { isAdmin } = useAuth();
  const [providers, setProviders] = useState<ProviderAdmin[]>([]);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const [providerDialogOpen, setProviderDialogOpen] = useState(false);
  const [providerDialogMode, setProviderDialogMode] =
    useState<"create" | "edit">("create");
  const [providerForm, setProviderForm] =
    useState<ProviderFormState>(defaultProviderForm);
  const [providerDialogError, setProviderDialogError] = useState<string | null>(
    null,
  );
  const [editingProvider, setEditingProvider] = useState<ProviderAdmin | null>(
    null,
  );

  const [modelDialogOpen, setModelDialogOpen] = useState(false);
  const [modelManagerProviderId, setModelManagerProviderId] = useState<
    string | null
  >(null);
  const [modelDialogMessage, setModelDialogMessage] = useState<string | null>(
    null,
  );
  const [modelDialogError, setModelDialogError] = useState<string | null>(null);

  const [modelFormOpen, setModelFormOpen] = useState(false);
  const [modelFormMode, setModelFormMode] = useState<"create" | "edit">(
    "create",
  );
  const [modelFormError, setModelFormError] = useState<string | null>(null);
  const [modelForm, setModelForm] =
    useState<ModelDialogForm>(defaultModelForm);
  const [editingModelId, setEditingModelId] = useState<string | null>(null);

  const modelManagerProvider = useMemo(
    () =>
      modelManagerProviderId
        ? providers.find((item) => item.id === modelManagerProviderId) ?? null
        : null,
    [modelManagerProviderId, providers],
  );

  const clearMessages = useCallback(() => {
    setError(null);
    setMessage(null);
  }, []);

  const loadProviders = useCallback(async () => {
    if (!isAdmin) {
      return;
    }
    setLoading(true);
    try {
      const data = await listProvidersAdmin();
      setProviders(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "加载厂商列表失败，请稍后再试",
      );
    } finally {
      setLoading(false);
    }
  }, [isAdmin]);

  useEffect(() => {
    void loadProviders();
  }, [loadProviders]);

  const handleRefresh = useCallback(() => {
    void loadProviders();
  }, [loadProviders]);

  const openCreateProviderDialog = useCallback(() => {
    setProviderDialogMode("create");
    setProviderForm(defaultProviderForm());
    setProviderDialogError(null);
    setEditingProvider(null);
    setProviderDialogOpen(true);
  }, []);

  const openEditProviderDialog = useCallback((provider: ProviderAdmin) => {
    setProviderDialogMode("edit");
    setProviderForm(toProviderFormState(provider));
    setProviderDialogError(null);
    setEditingProvider(provider);
    setProviderDialogOpen(true);
  }, []);

  const closeProviderDialog = useCallback(() => {
    if (busy) {
      return;
    }
    setProviderDialogOpen(false);
    setProviderDialogMode("create");
    setProviderForm(defaultProviderForm());
    setEditingProvider(null);
    setProviderDialogError(null);
  }, [busy]);

  const handleSubmitProvider = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setProviderDialogError(null);
      clearMessages();

      const trimmedName = providerForm.name.trim();
      const trimmedDriver = providerForm.driver.trim();
      if (!trimmedName || !trimmedDriver) {
        setProviderDialogError("厂商名称和驱动类型为必填项");
        return;
      }
      if (providerDialogMode === "create" && !providerForm.id.trim()) {
        setProviderDialogError("厂商 ID 为必填项");
        return;
      }

      const trimmedDescription = providerForm.description.trim();
      const trimmedBaseUrl = providerForm.base_url.trim();
      const trimmedConfig = providerForm.config_text.trim();

      let configObject: Record<string, unknown> | undefined;
      if (trimmedConfig) {
        try {
          configObject = JSON.parse(trimmedConfig);
        } catch {
          setProviderDialogError("配置 JSON 解析失败，请检查格式");
          return;
        }
      }

      setBusy(true);
      try {
        if (providerDialogMode === "create") {
          await createProvider({
            id: providerForm.id.trim(),
            name: trimmedName,
            driver: trimmedDriver,
            description: trimmedDescription || undefined,
            base_url: trimmedBaseUrl || undefined,
            api_key: providerForm.api_key.trim() || undefined,
            config: configObject,
            is_active: providerForm.is_active,
          });
          setMessage("厂商创建成功");
        } else if (editingProvider) {
          const payload: ProviderUpdatePayload = {};
          if (trimmedName !== editingProvider.name) {
            payload.name = trimmedName;
          }
          if (trimmedDriver !== editingProvider.driver) {
            payload.driver = trimmedDriver;
          }
          if (
            trimmedDescription !== (editingProvider.description ?? "")
          ) {
            payload.description = trimmedDescription;
          }
          if (trimmedBaseUrl !== (editingProvider.base_url ?? "")) {
            payload.base_url = trimmedBaseUrl;
          }
          const trimmedApiKey = providerForm.api_key.trim();
          if (trimmedApiKey) {
            payload.api_key = trimmedApiKey;
          }
          if (configObject) {
            payload.config = configObject;
          } else if (
            !trimmedConfig &&
            editingProvider.config &&
            Object.keys(editingProvider.config).length > 0
          ) {
            payload.config = {};
          }
          if (providerForm.is_active !== editingProvider.is_active) {
            payload.is_active = providerForm.is_active;
          }

          if (Object.keys(payload).length === 0) {
            setProviderDialogError("未检测到改动");
            setBusy(false);
            return;
          }

          await updateProvider(editingProvider.id, payload);
          setMessage("厂商信息已更新");
        }

        setProviderDialogOpen(false);
        setProviderDialogMode("create");
        setProviderForm(defaultProviderForm());
        setEditingProvider(null);
        await loadProviders();
      } catch (err) {
        setProviderDialogError(
          err instanceof Error ? err.message : "提交失败，请稍后再试",
        );
      } finally {
        setBusy(false);
      }
    },
    [
      clearMessages,
      editingProvider,
      loadProviders,
      providerDialogMode,
      providerForm,
    ],
  );

  const handleClearProviderKey = useCallback(async () => {
    if (!editingProvider) {
      return;
    }
    if (
      !window.confirm("确定要清空该厂商的 API 密钥吗？此操作将立即生效。")
    ) {
      return;
    }
    clearMessages();
    setProviderDialogError(null);
    setBusy(true);
    try {
      await updateProvider(editingProvider.id, { api_key: "" });
      setMessage("密钥已清空");
      setProviderDialogOpen(false);
      setProviderDialogMode("create");
      setProviderForm(defaultProviderForm());
      setEditingProvider(null);
      await loadProviders();
    } catch (err) {
      setProviderDialogError(
        err instanceof Error ? err.message : "清空密钥失败",
      );
    } finally {
      setBusy(false);
    }
  }, [clearMessages, editingProvider, loadProviders]);

  const handleToggleProviderActive = useCallback(
    async (provider: ProviderAdmin) => {
      clearMessages();
      setBusy(true);
      try {
        await updateProvider(provider.id, { is_active: !provider.is_active });
        setMessage("厂商状态已更新");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "更新厂商状态失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders],
  );

  const handleDeleteProvider = useCallback(
    async (provider: ProviderAdmin) => {
      if (
        !window.confirm("确认删除该厂商及其全部模型配置吗？此操作不可恢复。")
      ) {
        return;
      }
      clearMessages();
      setBusy(true);
      try {
        await deleteProvider(provider.id);
        setMessage("厂商已删除");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "删除厂商失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders],
  );

  const openModelManager = useCallback((provider: ProviderAdmin) => {
    setModelManagerProviderId(provider.id);
    setModelDialogMessage(null);
    setModelDialogError(null);
    setModelFormOpen(false);
    setModelForm(defaultModelForm());
    setModelFormMode("create");
    setEditingModelId(null);
    setModelFormError(null);
    setModelDialogOpen(true);
  }, []);

  const closeModelDialog = useCallback(() => {
    if (busy) {
      return;
    }
    setModelDialogOpen(false);
    setModelManagerProviderId(null);
    setModelDialogMessage(null);
    setModelDialogError(null);
    setModelFormOpen(false);
    setModelForm(defaultModelForm());
    setModelFormMode("create");
    setEditingModelId(null);
    setModelFormError(null);
  }, [busy]);

  const openCreateModelForm = useCallback(() => {
    if (!modelManagerProviderId) {
      return;
    }
    setModelForm(defaultModelForm());
    setModelFormMode("create");
    setModelFormError(null);
    setEditingModelId(null);
    setModelFormOpen(true);
  }, [modelManagerProviderId]);

  const openEditModelForm = useCallback((model: ProviderModelAdmin) => {
    setModelForm(toModelDialogForm(model));
    setModelFormMode("edit");
    setModelFormError(null);
    setEditingModelId(model.model_id);
    setModelFormOpen(true);
  }, []);

  const closeModelForm = useCallback(() => {
    if (busy) {
      return;
    }
    setModelFormOpen(false);
    setModelForm(defaultModelForm());
    setModelFormMode("create");
    setEditingModelId(null);
    setModelFormError(null);
  }, [busy]);

  const handleSubmitModelForm = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setModelFormError(null);
      setModelDialogError(null);
      setModelDialogMessage(null);
      clearMessages();

      if (!modelManagerProviderId) {
        setModelFormError("未找到厂商信息，请重试");
        return;
      }

      const provider =
        providers.find((item) => item.id === modelManagerProviderId) ?? null;
      if (!provider) {
        setModelFormError("未找到厂商信息，请重试");
        return;
      }

      const trimmedName = modelForm.name.trim();
      if (!trimmedName) {
        setModelFormError("模型名称为必填项");
        return;
      }

      const trimmedModelId = modelForm.model_id.trim();
      const trimmedDescription = modelForm.description.trim();
      const trimmedPrice = modelForm.price.trim();
      const trimmedSettings = modelForm.settings.trim();
      const trimmedMaxImages = modelForm.max_images.trim();

      let settingsObject: Record<string, unknown> | undefined;
      if (trimmedSettings) {
        try {
          settingsObject = JSON.parse(trimmedSettings);
        } catch {
          setModelFormError("模型配置 JSON 解析失败，请检查格式");
          return;
        }
      }

      let parsedMaxImages: number | undefined;
      if (trimmedMaxImages) {
        const parsed = Number.parseInt(trimmedMaxImages, 10);
        if (Number.isNaN(parsed) || parsed < 0) {
          setModelFormError("最大参考图数量需为非负整数");
          return;
        }
        parsedMaxImages = parsed;
      }

      const modalities = parseCommaSeparated(modelForm.modalities);
      const sizes = parseCommaSeparated(modelForm.supported_sizes);

      if (modelFormMode === "create") {
        if (!trimmedModelId) {
          setModelFormError("模型 ID 为必填项");
          return;
        }
        setBusy(true);
        try {
          await createProviderModel(provider.id, {
            model_id: trimmedModelId,
            name: trimmedName,
            description: trimmedDescription || undefined,
            price: trimmedPrice || undefined,
            max_images: parsedMaxImages,
            modalities: modalities.length > 0 ? modalities : undefined,
            supported_sizes: sizes.length > 0 ? sizes : undefined,
            settings: settingsObject,
            is_active: modelForm.is_active,
          });
          setModelDialogMessage("模型已创建");
          setModelFormOpen(false);
          setModelForm(defaultModelForm());
          await loadProviders();
        } catch (err) {
          setModelFormError(
            err instanceof Error ? err.message : "创建模型失败",
          );
        } finally {
          setBusy(false);
        }
        return;
      }

      if (!editingModelId) {
        setModelFormError("未找到模型信息，请重试");
        return;
      }
      const existingModel =
        provider.models?.find((item) => item.model_id === editingModelId) ??
        null;
      if (!existingModel) {
        setModelFormError("未找到模型信息，请重试");
        return;
      }

      const payload: ProviderModelUpdatePayload = {};
      if (trimmedName !== existingModel.name) {
        payload.name = trimmedName;
      }
      if (trimmedDescription !== (existingModel.description ?? "")) {
        payload.description = trimmedDescription;
      }
      if (trimmedPrice !== (existingModel.price ?? "")) {
        payload.price = trimmedPrice;
      }
      if (parsedMaxImages !== undefined) {
        if (parsedMaxImages !== existingModel.max_images) {
          payload.max_images = parsedMaxImages;
        }
      } else if (
        (existingModel.max_images ?? 0) !== 0 &&
        !trimmedMaxImages
      ) {
        payload.max_images = 0;
      }
      if (
        JSON.stringify(modalities) !==
        JSON.stringify(existingModel.modalities ?? [])
      ) {
        payload.modalities = modalities;
      }
      if (
        JSON.stringify(sizes) !==
        JSON.stringify(existingModel.supported_sizes ?? [])
      ) {
        payload.supported_sizes = sizes;
      }
      if (settingsObject) {
        payload.settings = settingsObject;
      } else if (
        !trimmedSettings &&
        existingModel.settings &&
        Object.keys(existingModel.settings).length > 0
      ) {
        payload.settings = {};
      }
      if (modelForm.is_active !== existingModel.is_active) {
        payload.is_active = modelForm.is_active;
      }

      if (Object.keys(payload).length === 0) {
        setModelFormError("未检测到改动");
        return;
      }

      setBusy(true);
      try {
        await updateProviderModel(provider.id, editingModelId, payload);
        setModelDialogMessage("模型信息已更新");
        setModelFormOpen(false);
        setModelForm(defaultModelForm());
        setEditingModelId(null);
        await loadProviders();
      } catch (err) {
        setModelFormError(
          err instanceof Error ? err.message : "更新模型失败",
        );
      } finally {
        setBusy(false);
      }
    },
    [
      clearMessages,
      editingModelId,
      loadProviders,
      modelForm,
      modelFormMode,
      modelManagerProviderId,
      providers,
    ],
  );

  const handleToggleModelActive = useCallback(
    async (model: ProviderModelAdmin) => {
      if (!modelManagerProviderId) {
        setModelDialogError("未找到厂商信息，请重试");
        return;
      }
      const provider =
        providers.find((item) => item.id === modelManagerProviderId) ?? null;
      if (!provider) {
        setModelDialogError("未找到厂商信息，请重试");
        return;
      }
      clearMessages();
      setModelDialogError(null);
      setBusy(true);
      try {
        await updateProviderModel(provider.id, model.model_id, {
          is_active: !model.is_active,
        });
        setModelDialogMessage("模型状态已更新");
        await loadProviders();
      } catch (err) {
        setModelDialogError(
          err instanceof Error ? err.message : "更新模型状态失败",
        );
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, modelManagerProviderId, providers],
  );

  const handleDeleteModel = useCallback(
    async (model: ProviderModelAdmin) => {
      if (!modelManagerProviderId) {
        setModelDialogError("未找到厂商信息，请重试");
        return;
      }
      const provider =
        providers.find((item) => item.id === modelManagerProviderId) ?? null;
      if (!provider) {
        setModelDialogError("未找到厂商信息，请重试");
        return;
      }
      if (
        !window.confirm(
          `确认删除模型 ${model.model_id} 吗？此操作不可恢复。`,
        )
      ) {
        return;
      }
      clearMessages();
      setModelDialogError(null);
      setBusy(true);
      try {
        await deleteProviderModel(provider.id, model.model_id);
        setModelDialogMessage("模型已删除");
        await loadProviders();
      } catch (err) {
        setModelDialogError(
          err instanceof Error ? err.message : "删除模型失败",
        );
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, modelManagerProviderId, providers],
  );

  if (!isAdmin) {
    return (
      <Box sx={{ py: 8 }}>
        <Paper sx={{ p: 6, textAlign: "center" }}>
          <Typography variant="h5" sx={{ fontWeight: 600, mb: 2 }}>
            厂商管理
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
          厂商管理
        </Typography>
        <Typography variant="body1" color="text.secondary">
          统一管理各类模型服务厂商及其模型配置，支持快速启用、停用与同步更新。
        </Typography>
      </Box>

      <Paper sx={{ p: 4, display: "flex", flexDirection: "column", gap: 3 }}>
        <Stack
          direction={{ xs: "column", sm: "row" }}
          spacing={2}
          justifyContent="space-between"
          alignItems={{ xs: "stretch", sm: "center" }}
        >
          <Box>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              厂商列表
            </Typography>
            <Typography variant="body2" color="text.secondary">
              支持按行操作模型管理、编辑、停用和删除。
            </Typography>
          </Box>
          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              onClick={handleRefresh}
              disabled={loading || busy}
            >
              刷新
            </Button>
            <Button
              variant="contained"
              startIcon={<AddRoundedIcon />}
              onClick={openCreateProviderDialog}
              disabled={busy}
            >
              新增厂商
            </Button>
          </Stack>
        </Stack>

        {(error || message) && (
          <Stack spacing={2}>
            {error && <Alert severity="error">{error}</Alert>}
            {message && <Alert severity="success">{message}</Alert>}
          </Stack>
        )}

        {loading ? (
          <Box
            sx={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              py: 6,
            }}
          >
            <CircularProgress />
          </Box>
        ) : providers.length > 0 ? (
          <Box sx={{ overflowX: "auto" }}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>厂商</TableCell>
                  <TableCell>驱动</TableCell>
                  <TableCell>状态</TableCell>
                  <TableCell>模型数量</TableCell>
                  <TableCell>描述</TableCell>
                  <TableCell align="right">操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {providers.map((provider) => (
                  <TableRow key={provider.id} hover>
                    <TableCell>
                      <Stack spacing={0.5}>
                        <Typography
                          variant="subtitle1"
                          sx={{ fontWeight: 600 }}
                        >
                          {provider.name}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {provider.id}
                        </Typography>
                      </Stack>
                    </TableCell>
                    <TableCell>{provider.driver}</TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        color={provider.is_active ? "success" : "default"}
                        label={provider.is_active ? "启用" : "停用"}
                      />
                    </TableCell>
                    <TableCell>{provider.models?.length ?? 0}</TableCell>
                    <TableCell sx={{ maxWidth: 240 }}>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        noWrap
                      >
                        {provider.description || "—"}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <Stack
                        direction="row"
                        spacing={1}
                        justifyContent="flex-end"
                      >
                        <Button
                          variant="outlined"
                          size="small"
                          onClick={() => openModelManager(provider)}
                          disabled={busy}
                        >
                          模型管理
                        </Button>
                        <Button
                          variant="outlined"
                          size="small"
                          onClick={() => openEditProviderDialog(provider)}
                          disabled={busy}
                        >
                          编辑
                        </Button>
                        <Button
                          variant="text"
                          size="small"
                          onClick={() =>
                            void handleToggleProviderActive(provider)
                          }
                          disabled={busy}
                        >
                          {provider.is_active ? "停用" : "启用"}
                        </Button>
                        <Button
                          variant="text"
                          size="small"
                          color="error"
                          startIcon={<DeleteForeverRoundedIcon fontSize="small" />}
                          onClick={() => void handleDeleteProvider(provider)}
                          disabled={busy}
                        >
                          删除
                        </Button>
                      </Stack>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        ) : (
          <Box
            sx={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              py: 6,
            }}
          >
            <Typography variant="body2" color="text.secondary">
              尚未配置任何厂商，请点击“新增厂商”完成首个配置。
            </Typography>
          </Box>
        )}
      </Paper>

      <Dialog
        open={providerDialogOpen}
        onClose={closeProviderDialog}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>
          {providerDialogMode === "create"
            ? "新增厂商"
            : `编辑厂商：${editingProvider?.name ?? ""}`}
        </DialogTitle>
        <Box component="form" onSubmit={handleSubmitProvider}>
          <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {providerDialogError && (
              <Alert severity="error">{providerDialogError}</Alert>
            )}
            {providerDialogMode === "create" && (
              <TextField
                label="厂商 ID"
                value={providerForm.id}
                onChange={(event) =>
                  setProviderForm((prev) => ({
                    ...prev,
                    id: event.target.value,
                  }))
                }
                required
                helperText="需要与后端持久化配置一致，建议使用小写字母与连字符"
              />
            )}
            <TextField
              label="显示名称"
              value={providerForm.name}
              onChange={(event) =>
                setProviderForm((prev) => ({
                  ...prev,
                  name: event.target.value,
                }))
              }
              required
            />
            <TextField
              label="驱动类型"
              value={providerForm.driver}
              onChange={(event) =>
                setProviderForm((prev) => ({
                  ...prev,
                  driver: event.target.value,
                }))
              }
              required
              helperText="例如 openrouter / dashscope / fal"
            />
            <TextField
              label="API 基础地址"
              value={providerForm.base_url}
              onChange={(event) =>
                setProviderForm((prev) => ({
                  ...prev,
                  base_url: event.target.value,
                }))
              }
              placeholder="可选，留空使用默认地址"
            />
            <TextField
              label="描述"
              value={providerForm.description}
              onChange={(event) =>
                setProviderForm((prev) => ({
                  ...prev,
                  description: event.target.value,
                }))
              }
              multiline
              minRows={2}
            />
            <TextField
              label="API 密钥"
              value={providerForm.api_key}
              onChange={(event) =>
                setProviderForm((prev) => ({
                  ...prev,
                  api_key: event.target.value,
                }))
              }
              placeholder={
                providerDialogMode === "edit"
                  ? "留空则保持不变"
                  : "可选，输入后立即生效"
              }
            />
            <TextField
              label="附加配置 (JSON)"
              value={providerForm.config_text}
              onChange={(event) =>
                setProviderForm((prev) => ({
                  ...prev,
                  config_text: event.target.value,
                }))
              }
              placeholder='例如 {"endpoint": "..."}'
              multiline
              minRows={3}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={providerForm.is_active}
                  onChange={(event) =>
                    setProviderForm((prev) => ({
                      ...prev,
                      is_active: event.target.checked,
                    }))
                  }
                />
              }
              label={providerForm.is_active ? "启用状态" : "停用状态"}
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3 }}>
            {providerDialogMode === "edit" && (
              <Button
                color="warning"
                onClick={handleClearProviderKey}
                startIcon={<RestartAltRoundedIcon />}
                disabled={busy}
              >
                清空密钥
              </Button>
            )}
            <Box sx={{ flexGrow: 1 }} />
            <Button onClick={closeProviderDialog} disabled={busy}>
              取消
            </Button>
            <Button type="submit" variant="contained" disabled={busy}>
              {busy
                ? "提交中..."
                : providerDialogMode === "create"
                  ? "创建"
                  : "保存"}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      <Dialog
        open={modelDialogOpen}
        onClose={closeModelDialog}
        fullWidth
        maxWidth="lg"
      >
        <DialogTitle>
          模型管理
          {modelManagerProvider ? `：${modelManagerProvider.name}` : ""}
        </DialogTitle>
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={2}
            justifyContent="space-between"
            alignItems={{ xs: "stretch", sm: "center" }}
          >
            <Box>
              <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                已配置模型
              </Typography>
              <Typography variant="body2" color="text.secondary">
                维护模型基础信息、开关状态及附加配置。
              </Typography>
            </Box>
            <Stack direction="row" spacing={1}>
              <Button
                variant="outlined"
                onClick={handleRefresh}
                disabled={loading || busy}
              >
                刷新
              </Button>
              <Button
                variant="contained"
                startIcon={<AddRoundedIcon />}
                onClick={openCreateModelForm}
                disabled={!modelManagerProviderId || busy}
              >
                新增模型
              </Button>
            </Stack>
          </Stack>

          {(modelDialogError || modelDialogMessage) && (
            <Stack spacing={2}>
              {modelDialogError && (
                <Alert severity="error">{modelDialogError}</Alert>
              )}
              {modelDialogMessage && (
                <Alert severity="success">{modelDialogMessage}</Alert>
              )}
            </Stack>
          )}

          {modelManagerProvider && modelManagerProvider.models?.length ? (
            <Box sx={{ overflowX: "auto" }}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>模型 ID</TableCell>
                    <TableCell>名称</TableCell>
                    <TableCell>价格/描述</TableCell>
                    <TableCell>状态</TableCell>
                    <TableCell align="right">操作</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {modelManagerProvider.models?.map((model) => (
                    <TableRow key={model.model_id} hover>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 600 }}>
                          {model.model_id}
                        </Typography>
                      </TableCell>
                      <TableCell>{model.name}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {model.price || model.description || "—"}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          size="small"
                          color={model.is_active ? "success" : "default"}
                          label={model.is_active ? "启用" : "停用"}
                        />
                      </TableCell>
                      <TableCell align="right">
                        <Stack
                          direction="row"
                          spacing={1}
                          justifyContent="flex-end"
                        >
                          <Button
                            variant="outlined"
                            size="small"
                            onClick={() => openEditModelForm(model)}
                            disabled={busy}
                          >
                            编辑
                          </Button>
                          <Button
                            variant="text"
                            size="small"
                            onClick={() =>
                              void handleToggleModelActive(model)
                            }
                            disabled={busy}
                          >
                            {model.is_active ? "停用" : "启用"}
                          </Button>
                          <Button
                            variant="text"
                            size="small"
                            color="error"
                            startIcon={
                              <DeleteForeverRoundedIcon fontSize="small" />
                            }
                            onClick={() => void handleDeleteModel(model)}
                            disabled={busy}
                          >
                            删除
                          </Button>
                        </Stack>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Box>
          ) : (
            <Typography variant="body2" color="text.secondary">
              暂无模型配置，可通过“新增模型”完成添加。
            </Typography>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button onClick={closeModelDialog} disabled={busy}>
            关闭
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={modelFormOpen}
        onClose={closeModelForm}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>
          {modelFormMode === "create"
            ? "新增模型"
            : `编辑模型：${modelForm.model_id}`}
        </DialogTitle>
        <Box component="form" onSubmit={handleSubmitModelForm}>
          <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {modelFormError && <Alert severity="error">{modelFormError}</Alert>}
            {modelFormMode === "create" ? (
              <TextField
                label="模型 ID"
                value={modelForm.model_id}
                onChange={(event) =>
                  setModelForm((prev) => ({
                    ...prev,
                    model_id: event.target.value,
                  }))
                }
                required
                helperText="需要与后端识别的模型标识一致"
              />
            ) : (
              <TextField
                label="模型 ID"
                value={modelForm.model_id}
                disabled
              />
            )}
            <TextField
              label="显示名称"
              value={modelForm.name}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  name: event.target.value,
                }))
              }
              required
            />
            <TextField
              label="价格 / 描述"
              value={modelForm.price}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  price: event.target.value,
                }))
              }
            />
            <TextField
              label="模型描述"
              value={modelForm.description}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  description: event.target.value,
                }))
              }
              multiline
              minRows={2}
            />
            <TextField
              label="最大参考图数量"
              value={modelForm.max_images}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  max_images: event.target.value,
                }))
              }
              placeholder="可选，需为非负整数"
            />
            <TextField
              label="支持模态（逗号分隔）"
              value={modelForm.modalities}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  modalities: event.target.value,
                }))
              }
              placeholder="例如 text, image"
            />
            <TextField
              label="支持尺寸（逗号分隔）"
              value={modelForm.supported_sizes}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  supported_sizes: event.target.value,
                }))
              }
              placeholder="例如 1K, 2K, 4K"
            />
            <TextField
              label="附加配置 (JSON)"
              value={modelForm.settings}
              onChange={(event) =>
                setModelForm((prev) => ({
                  ...prev,
                  settings: event.target.value,
                }))
              }
              placeholder='例如 {"endpoint": "/fal-ai/..."}'
              multiline
              minRows={3}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={modelForm.is_active}
                  onChange={(event) =>
                    setModelForm((prev) => ({
                      ...prev,
                      is_active: event.target.checked,
                    }))
                  }
                />
              }
              label={modelForm.is_active ? "启用状态" : "停用状态"}
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3 }}>
            <Button onClick={closeModelForm} disabled={busy}>
              取消
            </Button>
            <Button type="submit" variant="contained" disabled={busy}>
              {busy
                ? "提交中..."
                : modelFormMode === "create"
                  ? "创建"
                  : "保存"}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>
    </Box>
  );
};

export default ProviderManagementPage;
