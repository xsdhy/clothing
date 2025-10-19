import React, { useCallback, useEffect, useMemo, useState } from "react";
import Grid from "@mui/material/Grid";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  Paper,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import DeleteForeverRoundedIcon from "@mui/icons-material/DeleteForeverRounded";
import RestartAltRoundedIcon from "@mui/icons-material/RestartAltRounded";
import type {
  ProviderAdmin,
  ProviderModelAdmin,
  ProviderCreatePayload,
  ProviderUpdatePayload,
  ProviderModelCreatePayload,
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
  name: string;
  driver: string;
  description: string;
  base_url: string;
  api_key: string;
  config: string;
  is_active: boolean;
}

interface ModelFormState {
  name: string;
  description: string;
  price: string;
  max_images: string;
  modalities: string;
  supported_sizes: string;
  settings: string;
  is_active: boolean;
}

interface NewModelFormState extends ModelFormState {
  model_id: string;
}

const jsonStringify = (value?: Record<string, unknown>): string =>
  value && Object.keys(value).length > 0 ? JSON.stringify(value, null, 2) : "";

const toProviderForm = (provider: ProviderAdmin): ProviderFormState => ({
  name: provider.name,
  driver: provider.driver,
  description: provider.description ?? "",
  base_url: provider.base_url ?? "",
  api_key: "",
  config: jsonStringify(provider.config as Record<string, unknown>),
  is_active: provider.is_active,
});

const toModelForm = (model: ProviderModelAdmin): ModelFormState => ({
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

const emptyNewModelForm = (): NewModelFormState => ({
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

const parseCommaSeparated = (value: string): string[] =>
  value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => item.length > 0);

const ProviderManagementPage: React.FC = () => {
  const { isAdmin } = useAuth();
  const [providers, setProviders] = useState<ProviderAdmin[]>([]);
  const [providerForms, setProviderForms] = useState<
    Record<string, ProviderFormState>
  >({});
  const [modelForms, setModelForms] = useState<
    Record<string, Record<string, ModelFormState>>
  >({});
  const [newModelForms, setNewModelForms] = useState<
    Record<string, NewModelFormState>
  >({});
  const [newProvider, setNewProvider] = useState<
    ProviderCreatePayload & { config_text: string }
  >({
    id: "",
    name: "",
    driver: "",
    description: "",
    api_key: "",
    base_url: "",
    config_text: "",
    is_active: true,
  });
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const syncForms = useCallback((list: ProviderAdmin[]) => {
    const providerState: Record<string, ProviderFormState> = {};
    const modelState: Record<string, Record<string, ModelFormState>> = {};
    const newModelState: Record<string, NewModelFormState> = {};

    list.forEach((provider) => {
      providerState[provider.id] = toProviderForm(provider);
      const models = provider.models ?? [];
      const perProviderModels: Record<string, ModelFormState> = {};
      models.forEach((model) => {
        perProviderModels[model.model_id] = toModelForm(model);
      });
      modelState[provider.id] = perProviderModels;
      newModelState[provider.id] = emptyNewModelForm();
    });

    setProviderForms(providerState);
    setModelForms(modelState);
    setNewModelForms(newModelState);
  }, []);

  const loadProviders = useCallback(async () => {
    if (!isAdmin) {
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const data = await listProvidersAdmin();
      setProviders(data);
      syncForms(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "加载厂商列表失败，请稍后再试",
      );
    } finally {
      setLoading(false);
    }
  }, [isAdmin, syncForms]);

  useEffect(() => {
    void loadProviders();
  }, [loadProviders]);

  const providerHasData = useMemo(() => providers.length > 0, [providers]);

  const clearMessages = useCallback(() => {
    setError(null);
    setMessage(null);
  }, []);

  const handleCreateProvider = useCallback(
    async (event: React.FormEvent) => {
      event.preventDefault();
      clearMessages();
      const trimmedId = newProvider.id?.trim() ?? "";
      const trimmedName = newProvider.name?.trim() ?? "";
      const trimmedDriver = newProvider.driver?.trim() ?? "";
      if (!trimmedId || !trimmedName || !trimmedDriver) {
        setError("厂商 ID、名称和驱动类型为必填项");
        return;
      }

      let configObject: Record<string, unknown> | undefined;
      if (newProvider.config_text.trim()) {
        try {
          configObject = JSON.parse(newProvider.config_text);
        } catch (err) {
          setError("配置 JSON 解析失败，请检查格式");
          return;
        }
      }

      setBusy(true);
      try {
        await createProvider({
          id: trimmedId,
          name: trimmedName,
          driver: trimmedDriver,
          description: newProvider.description?.trim() || undefined,
          api_key: newProvider.api_key?.trim() || undefined,
          base_url: newProvider.base_url?.trim() || undefined,
          config: configObject,
          is_active: newProvider.is_active,
        });
        setMessage("厂商创建成功");
        setNewProvider({
          id: "",
          name: "",
          driver: "",
          description: "",
          api_key: "",
          base_url: "",
          config_text: "",
          is_active: true,
        });
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "创建厂商失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, newProvider],
  );

  const handleProviderFieldChange = useCallback(
    (
      providerId: string,
      field: keyof ProviderFormState,
      value: string | boolean,
    ) => {
      setProviderForms((prev) => ({
        ...prev,
        [providerId]: {
          ...prev[providerId],
          [field]: value as never,
        },
      }));
    },
    [],
  );

  const saveProviderChanges = useCallback(
    async (
      providerId: string,
      overridePayload?: Partial<ProviderUpdatePayload>,
    ) => {
      clearMessages();
      const provider = providers.find((item) => item.id === providerId);
      const form = providerForms[providerId];
      if (!provider || !form) {
        setError("未找到对应的厂商信息");
        return;
      }

      const payload: ProviderUpdatePayload = overridePayload
        ? { ...overridePayload }
        : {};
      if (!overridePayload) {
        if (form.name.trim() !== provider.name) {
          payload.name = form.name.trim();
        }
        if (form.driver.trim() !== provider.driver) {
          payload.driver = form.driver.trim();
        }
        if ((form.description ?? "").trim() !== (provider.description ?? "")) {
          payload.description = form.description.trim();
        }
        if ((form.base_url ?? "").trim() !== (provider.base_url ?? "")) {
          payload.base_url = form.base_url.trim();
        }
        if (form.api_key.trim()) {
          payload.api_key = form.api_key.trim();
        }
        if (form.config.trim()) {
          try {
            payload.config = JSON.parse(form.config);
          } catch {
            setError("配置 JSON 解析失败，请检查格式");
            return;
          }
        } else if (provider.config && Object.keys(provider.config).length > 0) {
          payload.config = {};
        }
        if (form.is_active !== provider.is_active) {
          payload.is_active = form.is_active;
        }

        if (Object.keys(payload).length === 0) {
          setMessage("没有检测到改动");
          return;
        }
      }

      setBusy(true);
      try {
        await updateProvider(providerId, payload);
        setMessage("厂商信息已更新");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "更新厂商失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, providerForms, providers],
  );

  const handleToggleProviderActive = useCallback(
    async (providerId: string) => {
      const provider = providers.find((item) => item.id === providerId);
      if (!provider) {
        setError("未找到对应的厂商信息");
        return;
      }
      await saveProviderChanges(providerId, {
        is_active: !provider.is_active,
      });
    },
    [providers, saveProviderChanges],
  );

  const handleClearProviderKey = useCallback(
    async (providerId: string) => {
      clearMessages();
      if (
        !window.confirm("确定要清空该厂商的 API 密钥吗？此操作将立即生效。")
      ) {
        return;
      }
      setBusy(true);
      try {
        await updateProvider(providerId, { api_key: "" });
        setMessage("密钥已清空");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "清空密钥失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders],
  );

  const handleDeleteProvider = useCallback(
    async (providerId: string) => {
      clearMessages();
      if (
        !window.confirm("确认删除该厂商及其全部模型配置吗？此操作不可恢复。")
      ) {
        return;
      }
      setBusy(true);
      try {
        await deleteProvider(providerId);
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

  const handleModelFieldChange = useCallback(
    (
      providerId: string,
      modelId: string,
      field: keyof ModelFormState,
      value: string | boolean,
    ) => {
      setModelForms((prev) => {
        const providerModels = { ...(prev[providerId] ?? {}) };
        providerModels[modelId] = {
          ...providerModels[modelId],
          [field]: value as never,
        };
        return {
          ...prev,
          [providerId]: providerModels,
        };
      });
    },
    [],
  );

  const handleNewModelFieldChange = useCallback(
    (
      providerId: string,
      field: keyof NewModelFormState,
      value: string | boolean,
    ) => {
      setNewModelForms((prev) => ({
        ...prev,
        [providerId]: {
          ...(prev[providerId] ?? emptyNewModelForm()),
          [field]: value as never,
        },
      }));
    },
    [],
  );

  const saveModelChanges = useCallback(
    async (
      providerId: string,
      modelId: string,
      overridePayload?: Partial<ProviderModelUpdatePayload>,
    ) => {
      clearMessages();
      const provider = providers.find((item) => item.id === providerId);
      const model = provider?.models?.find((item) => item.model_id === modelId);
      const form = modelForms[providerId]?.[modelId];
      if (!provider || !model || !form) {
        setError("未找到对应的模型信息");
        return;
      }

      const payload: ProviderModelUpdatePayload = overridePayload
        ? { ...overridePayload }
        : {};
      if (!overridePayload) {
        if (form.name.trim() !== model.name) {
          payload.name = form.name.trim();
        }
        if ((form.description ?? "").trim() !== (model.description ?? "")) {
          payload.description = form.description.trim();
        }
        if ((form.price ?? "").trim() !== (model.price ?? "")) {
          payload.price = form.price.trim();
        }
        if (form.max_images.trim()) {
          const parsed = Number.parseInt(form.max_images.trim(), 10);
          if (Number.isNaN(parsed) || parsed < 0) {
            setError("最大参考图数量需为非负整数");
            return;
          }
          payload.max_images = parsed;
        } else if (model.max_images && model.max_images !== 0) {
          payload.max_images = 0;
        }
        const modalities = parseCommaSeparated(form.modalities);
        if (
          JSON.stringify(modalities) !==
          JSON.stringify(model.modalities ?? [])
        ) {
          payload.modalities = modalities;
        }
        const sizes = parseCommaSeparated(form.supported_sizes);
        if (
          JSON.stringify(sizes) !==
          JSON.stringify(model.supported_sizes ?? [])
        ) {
          payload.supported_sizes = sizes;
        }
        if (form.settings.trim()) {
          try {
            payload.settings = JSON.parse(form.settings);
          } catch {
            setError("模型配置 JSON 解析失败，请检查格式");
            return;
          }
        } else if (model.settings && Object.keys(model.settings).length > 0) {
          payload.settings = {};
        }
        if (form.is_active !== model.is_active) {
          payload.is_active = form.is_active;
        }

        if (Object.keys(payload).length === 0) {
          setMessage("没有检测到模型改动");
          return;
        }
      }

      setBusy(true);
      try {
        await updateProviderModel(providerId, modelId, payload);
        setMessage("模型信息已更新");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "更新模型失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, modelForms, providers],
  );

  const handleToggleModelActive = useCallback(
    async (providerId: string, modelId: string) => {
      const provider = providers.find((item) => item.id === providerId);
      const model = provider?.models?.find((item) => item.model_id === modelId);
      if (!model) {
        setError("未找到对应的模型信息");
        return;
      }
      await saveModelChanges(providerId, modelId, {
        is_active: !model.is_active,
      });
    },
    [providers, saveModelChanges],
  );

  const handleDeleteModel = useCallback(
    async (providerId: string, modelId: string) => {
      clearMessages();
      if (!window.confirm("确认删除该模型配置吗？")) {
        return;
      }
      setBusy(true);
      try {
        await deleteProviderModel(providerId, modelId);
        setMessage("模型已删除");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "删除模型失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders],
  );

  const handleCreateModel = useCallback(
    async (providerId: string) => {
      clearMessages();
      const form = newModelForms[providerId] ?? emptyNewModelForm();
      const trimmedModelId = form.model_id.trim();
      const trimmedName = form.name.trim();
      if (!trimmedModelId || !trimmedName) {
        setError("请填写模型 ID 与名称");
        return;
      }

      const payload: ProviderModelCreatePayload = {
        model_id: trimmedModelId,
        name: trimmedName,
        description: form.description.trim() || undefined,
        price: form.price.trim() || undefined,
        modalities: parseCommaSeparated(form.modalities),
        supported_sizes: parseCommaSeparated(form.supported_sizes),
        is_active: form.is_active,
      };

      if (form.max_images.trim()) {
        const parsed = Number.parseInt(form.max_images.trim(), 10);
        if (Number.isNaN(parsed) || parsed < 0) {
          setError("最大参考图数量需为非负整数");
          return;
        }
        payload.max_images = parsed;
      }

      let settingsObject: Record<string, unknown> | undefined;
      if (form.settings.trim()) {
        try {
          settingsObject = JSON.parse(form.settings);
        } catch {
          setError("模型配置 JSON 解析失败，请检查格式");
          return;
        }
      }
      if (settingsObject) {
        payload.settings = settingsObject;
      }

      setBusy(true);
      try {
        await createProviderModel(providerId, payload);
        setMessage("模型已创建");
        setNewModelForms((prev) => ({
          ...prev,
          [providerId]: emptyNewModelForm(),
        }));
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "创建模型失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, newModelForms],
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
          厂商与模型管理
        </Typography>
        <Typography variant="body1" color="text.secondary">
          在此集中维护厂商凭证及其模型配置，所有改动即时生效。
        </Typography>
      </Box>

      {(error || message) && (
        <Stack spacing={2}>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}
          {message && (
            <Alert severity="success" onClose={() => setMessage(null)}>
              {message}
            </Alert>
          )}
        </Stack>
      )}

      <Paper
        component="form"
        onSubmit={handleCreateProvider}
        sx={{ p: 4, display: "flex", flexDirection: "column", gap: 3 }}
      >
        <Box>
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            创建新厂商
          </Typography>
          <Typography variant="body2" color="text.secondary">
            厂商 ID 将用于调用接口及记录追踪，建议使用全小写英文字母。
          </Typography>
        </Box>
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 4 }}>
            <TextField
              label="厂商 ID"
              value={newProvider.id}
              required
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  id: event.target.value,
                }))
              }
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            <TextField
              label="显示名称"
              value={newProvider.name}
              required
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  name: event.target.value,
                }))
              }
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            <TextField
              label="驱动类型"
              value={newProvider.driver}
              required
              helperText="例如 openrouter / dashscope / fal"
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  driver: event.target.value,
                }))
              }
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 12, md: 6 }}>
            <TextField
              label="API 基础地址"
              value={newProvider.base_url}
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  base_url: event.target.value,
                }))
              }
              placeholder="可选，留空使用默认地址"
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 12, md: 6 }}>
            <TextField
              label="API 密钥"
              value={newProvider.api_key}
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  api_key: event.target.value,
                }))
              }
              placeholder="输入后立即生效"
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 12 }}>
            <TextField
              label="描述"
              value={newProvider.description}
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  description: event.target.value,
                }))
              }
              multiline
              minRows={2}
              fullWidth
            />
          </Grid>
          <Grid size={{ xs: 12 }}>
            <TextField
              label="附加配置 (JSON)"
              value={newProvider.config_text}
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  config_text: event.target.value,
                }))
              }
              placeholder='例如 {"endpoint": "..."}'
              multiline
              minRows={3}
              fullWidth
            />
          </Grid>
        </Grid>
        <Stack
          direction="row"
          spacing={2}
          alignItems="center"
          justifyContent="space-between"
        >
          <Stack direction="row" spacing={1} alignItems="center">
            <Switch
              checked={newProvider.is_active ?? true}
              onChange={(event) =>
                setNewProvider((prev) => ({
                  ...prev,
                  is_active: event.target.checked,
                }))
              }
            />
            <Typography variant="body2">
              {newProvider.is_active ? "启用状态" : "停用状态"}
            </Typography>
          </Stack>
          <Button
            type="submit"
            variant="contained"
            disabled={busy}
            sx={{ minWidth: 160 }}
          >
            创建厂商
          </Button>
        </Stack>
      </Paper>

      {loading ? (
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            py: 8,
          }}
        >
          <CircularProgress />
        </Box>
      ) : providerHasData ? (
        <Stack spacing={4}>
          {providers.map((provider) => {
            const form = providerForms[provider.id] ?? toProviderForm(provider);
            const models = provider.models ?? [];
            return (
              <Paper key={provider.id} sx={{ p: { xs: 3, md: 4 } }}>
                <Stack spacing={3}>
                  <Box
                    sx={{
                      display: "flex",
                      flexDirection: { xs: "column", md: "row" },
                      alignItems: { xs: "flex-start", md: "center" },
                      justifyContent: "space-between",
                      gap: 2,
                    }}
                  >
                    <Box>
                      <Typography variant="h5" sx={{ fontWeight: 700 }}>
                        {provider.name}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        驱动：{provider.driver} / ID：{provider.id}
                      </Typography>
                    </Box>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Chip
                        label={provider.is_active ? "已启用" : "已停用"}
                        color={provider.is_active ? "success" : "default"}
                      />
                      <Chip
                        label={provider.has_api_key ? "已配置密钥" : "未配置密钥"}
                        color={provider.has_api_key ? "primary" : "default"}
                      />
                    </Stack>
                  </Box>

                  <Grid container spacing={3}>
                    <Grid size={{ xs: 12, md: 6 }}>
                      <Stack spacing={3}>
                        <TextField
                          label="显示名称"
                          value={form.name}
                          onChange={(event) =>
                            handleProviderFieldChange(
                              provider.id,
                              "name",
                              event.target.value,
                            )
                          }
                          fullWidth
                        />
                        <TextField
                          label="驱动类型"
                          helperText="需与后端驱动名称一致"
                          value={form.driver}
                          onChange={(event) =>
                            handleProviderFieldChange(
                              provider.id,
                              "driver",
                              event.target.value,
                            )
                          }
                          fullWidth
                        />
                        <TextField
                          label="API 基础地址"
                          value={form.base_url}
                          onChange={(event) =>
                            handleProviderFieldChange(
                              provider.id,
                              "base_url",
                              event.target.value,
                            )
                          }
                          placeholder="留空使用默认地址"
                          fullWidth
                        />
                        <TextField
                          label="描述"
                          value={form.description}
                          onChange={(event) =>
                            handleProviderFieldChange(
                              provider.id,
                              "description",
                              event.target.value,
                            )
                          }
                          multiline
                          minRows={2}
                          fullWidth
                        />
                      </Stack>
                    </Grid>
                    <Grid size={{ xs: 12, md: 6 }}>
                      <Stack spacing={3}>
                        <TextField
                          label="API 密钥"
                          value={form.api_key}
                          onChange={(event) =>
                            handleProviderFieldChange(
                              provider.id,
                              "api_key",
                              event.target.value,
                            )
                          }
                          placeholder="输入新值以更新，留空不修改"
                          fullWidth
                        />
                        <TextField
                          label="附加配置 (JSON)"
                          value={form.config}
                          onChange={(event) =>
                            handleProviderFieldChange(
                              provider.id,
                              "config",
                              event.target.value,
                            )
                          }
                          placeholder='例如 {"endpoint": "..."}'
                          multiline
                          minRows={4}
                          fullWidth
                        />
                        <Stack direction="row" spacing={1} alignItems="center">
                          <Switch
                            checked={form.is_active}
                            onChange={(event) =>
                              handleProviderFieldChange(
                                provider.id,
                                "is_active",
                                event.target.checked,
                              )
                            }
                          />
                          <Typography variant="body2">
                            {form.is_active ? "启用状态" : "停用状态"}
                          </Typography>
                        </Stack>
                      </Stack>
                    </Grid>
                  </Grid>

                  <Stack
                    direction={{ xs: "column", md: "row" }}
                    spacing={2}
                    justifyContent="flex-end"
                  >
                    <Button
                      variant="contained"
                      onClick={() => saveProviderChanges(provider.id)}
                      disabled={busy}
                    >
                      保存修改
                    </Button>
                    <Button
                      variant="outlined"
                      color="warning"
                      startIcon={<RestartAltRoundedIcon />}
                      onClick={() => handleClearProviderKey(provider.id)}
                      disabled={busy || !provider.has_api_key}
                    >
                      清空密钥
                    </Button>
                    <Button
                      variant="outlined"
                      color={provider.is_active ? "secondary" : "success"}
                      onClick={() => handleToggleProviderActive(provider.id)}
                      disabled={busy}
                    >
                      {provider.is_active ? "停用厂商" : "启用厂商"}
                    </Button>
                    <Button
                      variant="outlined"
                      color="error"
                      startIcon={<DeleteForeverRoundedIcon />}
                      onClick={() => handleDeleteProvider(provider.id)}
                      disabled={busy}
                    >
                      删除厂商
                    </Button>
                  </Stack>

                  <Divider />

                  <Stack spacing={2}>
                    <Typography variant="h6" sx={{ fontWeight: 600 }}>
                      模型配置
                    </Typography>
                    {models.length === 0 ? (
                      <Typography variant="body2" color="text.secondary">
                        该厂商暂未配置模型，添加后即可在前端选择。
                      </Typography>
                    ) : (
                      <Stack spacing={2}>
                        {models.map((model) => {
                          const modelForm =
                            modelForms[provider.id]?.[model.model_id] ??
                            toModelForm(model);
                          return (
                            <Card key={model.model_id} variant="outlined">
                              <CardContent>
                                <Stack spacing={2}>
                                  <Box
                                    sx={{
                                      display: "flex",
                                      flexDirection: {
                                        xs: "column",
                                        md: "row",
                                      },
                                      alignItems: {
                                        xs: "flex-start",
                                        md: "center",
                                      },
                                      justifyContent: "space-between",
                                      gap: 1.5,
                                    }}
                                  >
                                    <Typography
                                      variant="subtitle1"
                                      sx={{ fontWeight: 600 }}
                                    >
                                      {model.name}（{model.model_id}）
                                    </Typography>
                                    <Stack
                                      direction="row"
                                      spacing={1}
                                      alignItems="center"
                                    >
                                      <Chip
                                        label={
                                          modelForm.is_active
                                            ? "已启用"
                                            : "已停用"
                                        }
                                        size="small"
                                        color={
                                          modelForm.is_active
                                            ? "success"
                                            : "default"
                                        }
                                      />
                                      {model.price && (
                                        <Chip
                                          label={`价格: ${model.price}`}
                                          size="small"
                                          color="primary"
                                          variant="outlined"
                                        />
                                      )}
                                    </Stack>
                                  </Box>

                                  <Grid container spacing={2}>
                                    <Grid size={{ xs: 12, md: 4 }}>
                                      <TextField
                                        label="显示名称"
                                        value={modelForm.name}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "name",
                                            event.target.value,
                                          )
                                        }
                                        fullWidth
                                      />
                                    </Grid>
                                    <Grid size={{ xs: 12, md: 4 }}>
                                      <TextField
                                        label="价格 / 描述"
                                        value={modelForm.price}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "price",
                                            event.target.value,
                                          )
                                        }
                                        placeholder="可选，例如 $0.02"
                                        fullWidth
                                      />
                                    </Grid>
                                    <Grid size={{ xs: 12, md: 4 }}>
                                      <TextField
                                        label="最大参考图数量"
                                        value={modelForm.max_images}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "max_images",
                                            event.target.value,
                                          )
                                        }
                                        placeholder="留空则不限"
                                        fullWidth
                                      />
                                    </Grid>
                                    <Grid size={{ xs: 12, md: 6 }}>
                                      <TextField
                                        label="支持模态（逗号分隔）"
                                        value={modelForm.modalities}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "modalities",
                                            event.target.value,
                                          )
                                        }
                                        placeholder="例如 text, image"
                                        fullWidth
                                      />
                                    </Grid>
                                    <Grid size={{ xs: 12, md: 6 }}>
                                      <TextField
                                        label="支持尺寸（逗号分隔）"
                                        value={modelForm.supported_sizes}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "supported_sizes",
                                            event.target.value,
                                          )
                                        }
                                        placeholder="例如 1K, 2K, 4K"
                                        fullWidth
                                      />
                                    </Grid>
                                    <Grid size={{ xs: 12 }}>
                                      <TextField
                                        label="描述"
                                        value={modelForm.description}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "description",
                                            event.target.value,
                                          )
                                        }
                                        multiline
                                        minRows={2}
                                        fullWidth
                                      />
                                    </Grid>
                                    <Grid size={{ xs: 12 }}>
                                      <TextField
                                        label="附加配置 (JSON)"
                                        value={modelForm.settings}
                                        onChange={(event) =>
                                          handleModelFieldChange(
                                            provider.id,
                                            model.model_id,
                                            "settings",
                                            event.target.value,
                                          )
                                        }
                                        placeholder='例如 {"endpoint": "/fal-ai/..."}'
                                        multiline
                                        minRows={3}
                                        fullWidth
                                      />
                                    </Grid>
                                  </Grid>

                                  <Stack
                                    direction={{ xs: "column", md: "row" }}
                                    spacing={2}
                                    justifyContent="flex-end"
                                  >
                                    <Button
                                      variant="contained"
                                      onClick={() =>
                                        saveModelChanges(
                                          provider.id,
                                          model.model_id,
                                        )
                                      }
                                      disabled={busy}
                                    >
                                      保存模型
                                    </Button>
                                    <Button
                                      variant="outlined"
                                      color={
                                        modelForm.is_active
                                          ? "secondary"
                                          : "success"
                                      }
                                      onClick={() =>
                                        handleToggleModelActive(
                                          provider.id,
                                          model.model_id,
                                        )
                                      }
                                      disabled={busy}
                                    >
                                      {modelForm.is_active ? "停用模型" : "启用模型"}
                                    </Button>
                                    <Button
                                      variant="outlined"
                                      color="error"
                                      startIcon={<DeleteForeverRoundedIcon />}
                                      onClick={() =>
                                        handleDeleteModel(
                                          provider.id,
                                          model.model_id,
                                        )
                                      }
                                      disabled={busy}
                                    >
                                      删除模型
                                    </Button>
                                  </Stack>
                                </Stack>
                              </CardContent>
                            </Card>
                          );
                        })}
                      </Stack>
                    )}

                    <Card variant="outlined">
                      <CardContent>
                        <Stack
                          component="form"
                          spacing={2}
                          onSubmit={(event) => {
                            event.preventDefault();
                            void handleCreateModel(provider.id);
                          }}
                        >
                          <Typography
                            variant="subtitle1"
                            sx={{ fontWeight: 600 }}
                          >
                            新增模型
                          </Typography>
                          <Grid container spacing={2}>
                            <Grid size={{ xs: 12, md: 4 }}>
                              <TextField
                                label="模型 ID"
                                value={
                                  newModelForms[provider.id]?.model_id ?? ""
                                }
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "model_id",
                                    event.target.value,
                                  )
                                }
                                placeholder="必须与后端模型标识一致"
                                required
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4 }}>
                              <TextField
                                label="显示名称"
                                value={newModelForms[provider.id]?.name ?? ""}
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "name",
                                    event.target.value,
                                  )
                                }
                                required
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12, md: 4 }}>
                              <TextField
                                label="价格 / 描述"
                                value={newModelForms[provider.id]?.price ?? ""}
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "price",
                                    event.target.value,
                                  )
                                }
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12, md: 6 }}>
                              <TextField
                                label="最大参考图数量"
                                value={
                                  newModelForms[provider.id]?.max_images ?? ""
                                }
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "max_images",
                                    event.target.value,
                                  )
                                }
                                placeholder="可选"
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12, md: 6 }}>
                              <TextField
                                label="支持模态（逗号分隔）"
                                value={
                                  newModelForms[provider.id]?.modalities ?? ""
                                }
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "modalities",
                                    event.target.value,
                                  )
                                }
                                placeholder="例如 text, image"
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                              <TextField
                                label="支持尺寸（逗号分隔）"
                                value={
                                  newModelForms[provider.id]
                                    ?.supported_sizes ?? ""
                                }
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "supported_sizes",
                                    event.target.value,
                                  )
                                }
                                placeholder="例如 1K, 2K, 4K"
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                              <TextField
                                label="模型描述"
                                value={
                                  newModelForms[provider.id]?.description ?? ""
                                }
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "description",
                                    event.target.value,
                                  )
                                }
                                multiline
                                minRows={2}
                                fullWidth
                              />
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                              <TextField
                                label="附加配置 (JSON)"
                                value={newModelForms[provider.id]?.settings ?? ""}
                                onChange={(event) =>
                                  handleNewModelFieldChange(
                                    provider.id,
                                    "settings",
                                    event.target.value,
                                  )
                                }
                                placeholder='例如 {"endpoint": "/fal-ai/..."}'
                                multiline
                                minRows={3}
                                fullWidth
                              />
                            </Grid>
                          </Grid>
                          <Stack
                            direction="row"
                            spacing={1}
                            alignItems="center"
                          >
                            <Switch
                              checked={
                                newModelForms[provider.id]?.is_active ?? true
                              }
                              onChange={(event) =>
                                handleNewModelFieldChange(
                                  provider.id,
                                  "is_active",
                                  event.target.checked,
                                )
                              }
                            />
                            <Typography variant="body2">
                              {newModelForms[provider.id]?.is_active ?? true
                                ? "启用状态"
                                : "停用状态"}
                            </Typography>
                          </Stack>
                          <Button
                            type="submit"
                            variant="contained"
                            disabled={busy}
                            sx={{ alignSelf: "flex-end" }}
                          >
                            添加模型
                          </Button>
                        </Stack>
                      </CardContent>
                    </Card>
                  </Stack>
                </Stack>
              </Paper>
            );
          })}
        </Stack>
      ) : (
        <Card>
          <CardContent>
            <Typography variant="body1" color="text.secondary">
              尚未配置任何厂商，请使用上方表单创建新厂商。
            </Typography>
          </CardContent>
        </Card>
      )}
    </Box>
  );
};

export default ProviderManagementPage;
