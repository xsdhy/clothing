import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
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
// Import new sub-components
import ProviderSidebar from "../components/provider/ProviderSidebar";
import ProviderDetailPanel from "../components/provider/ProviderDetailPanel";

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
  input_modalities: string[];
  output_modalities: string[];
  supported_sizes: string;
  supported_durations: string;
  default_size: string;
  default_duration: string;
  settings: string;
  // New fields for model capabilities
  generation_mode: string;
  endpoint_path: string;
  supports_streaming: boolean;
  supports_cancel: boolean;
  is_active: boolean;
}

const jsonStringify = (value?: Record<string, unknown>): string =>
  value && Object.keys(value).length > 0 ? JSON.stringify(value, null, 2) : "";

const parseCommaSeparated = (value: string): string[] =>
  value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => item.length > 0);

const parseNumberList = (value: string): number[] => {
  const seen = new Set<number>();
  return value
    .split(/,|，/)
    .map((item) => item.trim())
    .map((item) => Number.parseInt(item, 10))
    .filter((num) => Number.isFinite(num) && num > 0 && !seen.has(num))
    .map((num) => {
      seen.add(num);
      return num;
    });
};

const parsePositiveInt = (value: string): number | undefined => {
  const parsed = Number.parseInt(value.trim(), 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return undefined;
  }
  return parsed;
};

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
  input_modalities: [],
  output_modalities: [],
  supported_sizes: "",
  supported_durations: "",
  default_size: "",
  default_duration: "",
  settings: "",
  generation_mode: "",
  endpoint_path: "",
  supports_streaming: false,
  supports_cancel: false,
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
  input_modalities: model.input_modalities ?? [],
  output_modalities: model.output_modalities ?? [],
  supported_sizes: model.supported_sizes?.join(", ") ?? "",
  supported_durations:
    model.supported_durations && model.supported_durations.length
      ? model.supported_durations.join(", ")
      : "",
  default_size: model.default_size ?? "",
  default_duration:
    typeof model.default_duration === "number" && model.default_duration > 0
      ? String(model.default_duration)
      : "",
  settings: jsonStringify(model.settings as Record<string, unknown>),
  generation_mode: model.generation_mode ?? "",
  endpoint_path: model.endpoint_path ?? "",
  supports_streaming: model.supports_streaming ?? false,
  supports_cancel: model.supports_cancel ?? false,
  is_active: model.is_active,
});

const modalityOptions = ["text", "image", "video"];
const generationModeOptions = [
  "text_to_image",
  "image_to_image",
  "text_to_video",
  "image_to_video",
];

const normalizeMultiSelect = (value: string | string[]): string[] =>
  Array.isArray(value)
    ? value
    : value
        .split(",")
        .map((item) => item.trim())
        .filter((item) => item.length > 0);

const ProviderManagementPage: React.FC = () => {
  const { isAdmin } = useAuth();
  const [providers, setProviders] = useState<ProviderAdmin[]>([]);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  // New state for split view
  const [selectedProviderId, setSelectedProviderId] = useState<string | null>(
    null
  );

  const selectedProvider = useMemo(
    () => providers.find((p) => p.id === selectedProviderId) ?? null,
    [providers, selectedProviderId]
  );

  const [providerDialogOpen, setProviderDialogOpen] = useState(false);
  const [providerDialogMode, setProviderDialogMode] =
    useState<"create" | "edit">("create");
  const [providerForm, setProviderForm] =
    useState<ProviderFormState>(defaultProviderForm);
  const [providerDialogError, setProviderDialogError] = useState<string | null>(
    null
  );
  const [editingProvider, setEditingProvider] = useState<ProviderAdmin | null>(
    null
  );

  const [modelFormOpen, setModelFormOpen] = useState(false);
  const [modelFormMode, setModelFormMode] = useState<"create" | "edit">(
    "create"
  );
  const [modelFormError, setModelFormError] = useState<string | null>(null);
  const [modelForm, setModelForm] =
    useState<ModelDialogForm>(defaultModelForm);
  const [modelFormProviderId, setModelFormProviderId] = useState<string | null>(
    null
  );
  const [editingModelId, setEditingModelId] = useState<string | null>(null);

  const modelFormProvider = useMemo(
    () =>
      modelFormProviderId
        ? providers.find((item) => item.id === modelFormProviderId) ?? null
        : null,
    [modelFormProviderId, providers]
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
        err instanceof Error ? err.message : "加载厂商列表失败，请稍后再试"
      );
    } finally {
      setLoading(false);
    }
  }, [isAdmin]);

  useEffect(() => {
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
          // Select the new provider
          setSelectedProviderId(providerForm.id.trim());
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
          err instanceof Error ? err.message : "提交失败，请稍后再试"
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
    ]
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
        err instanceof Error ? err.message : "清空密钥失败"
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
    [clearMessages, loadProviders]
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
        if (selectedProviderId === provider.id) {
          setSelectedProviderId(null);
        }
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "删除厂商失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, selectedProviderId]
  );

  const openCreateModelForm = useCallback(
    (providerId: string) => {
      clearMessages();
      setModelForm(defaultModelForm());
      setModelFormMode("create");
      setModelFormError(null);
      setEditingModelId(null);
      setModelFormProviderId(providerId);
      setModelFormOpen(true);
    },
    [clearMessages]
  );

  const openEditModelForm = useCallback(
    (providerId: string, model: ProviderModelAdmin) => {
      clearMessages();
      setModelForm(toModelDialogForm(model));
      setModelFormMode("edit");
      setModelFormError(null);
      setEditingModelId(model.model_id);
      setModelFormProviderId(providerId);
      setModelFormOpen(true);
    },
    [clearMessages]
  );

  const handleCloneModel = useCallback(
    (providerId: string, model: ProviderModelAdmin) => {
      clearMessages();
      const clonedForm = toModelDialogForm(model);
      // Reset ID for clone
      clonedForm.model_id = "";
      setModelForm(clonedForm);
      setModelFormMode("create");
      setModelFormError(null);
      setEditingModelId(null);
      setModelFormProviderId(providerId);
      setModelFormOpen(true);
      setMessage(`正在克隆模型: ${model.name}`);
    },
    [clearMessages]
  );

  const closeModelForm = useCallback(() => {
    if (busy) {
      return;
    }
    setModelFormOpen(false);
    setModelForm(defaultModelForm());
    setModelFormMode("create");
    setEditingModelId(null);
    setModelFormError(null);
    setModelFormProviderId(null);
  }, [busy]);

  const handleSubmitModelForm = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setModelFormError(null);
      clearMessages();

      if (!modelFormProviderId) {
        setModelFormError("未找到厂商信息，请重试");
        return;
      }

      const provider =
        providers.find((item) => item.id === modelFormProviderId) ?? null;
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
      const trimmedDefaultSize = modelForm.default_size.trim();

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

      const inputModalities = modelForm.input_modalities;
      const outputModalities = modelForm.output_modalities;
      const sizes = parseCommaSeparated(modelForm.supported_sizes);
      const durations = parseNumberList(modelForm.supported_durations);
      const parsedDefaultDuration = parsePositiveInt(modelForm.default_duration);

      if (
        trimmedDefaultSize &&
        sizes.length > 0 &&
        !sizes.includes(trimmedDefaultSize)
      ) {
        setModelFormError("默认尺寸需包含在支持尺寸列表中");
        return;
      }

      if (
        parsedDefaultDuration !== undefined &&
        durations.length > 0 &&
        !durations.includes(parsedDefaultDuration)
      ) {
        setModelFormError("默认视频时长需包含在支持时长列表中");
        return;
      }

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
            input_modalities: inputModalities.length > 0 ? inputModalities : undefined,
            output_modalities:
              outputModalities.length > 0 ? outputModalities : undefined,
            supported_sizes: sizes.length > 0 ? sizes : undefined,
            supported_durations: durations.length > 0 ? durations : undefined,
            default_size: trimmedDefaultSize || undefined,
            default_duration: parsedDefaultDuration,
            settings: settingsObject,
            generation_mode: modelForm.generation_mode.trim() || undefined,
            endpoint_path: modelForm.endpoint_path.trim() || undefined,
            supports_streaming: modelForm.supports_streaming,
            supports_cancel: modelForm.supports_cancel,
            is_active: modelForm.is_active,
          });
          setMessage("模型已创建");
          setModelFormOpen(false);
          setModelForm(defaultModelForm());
          setModelFormProviderId(null);
          await loadProviders();
        } catch (err) {
          setModelFormError(
            err instanceof Error ? err.message : "创建模型失败"
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
        JSON.stringify(inputModalities) !==
        JSON.stringify(existingModel.input_modalities ?? [])
      ) {
        payload.input_modalities = inputModalities;
      }
      if (
        JSON.stringify(outputModalities) !==
        JSON.stringify(existingModel.output_modalities ?? [])
      ) {
        payload.output_modalities = outputModalities;
      }
      if (
        JSON.stringify(sizes) !==
        JSON.stringify(existingModel.supported_sizes ?? [])
      ) {
        payload.supported_sizes = sizes;
      }
      if (
        JSON.stringify(durations) !==
        JSON.stringify(existingModel.supported_durations ?? [])
      ) {
        payload.supported_durations = durations;
      }
      if (trimmedDefaultSize !== (existingModel.default_size ?? "")) {
        payload.default_size = trimmedDefaultSize;
      }
      if (parsedDefaultDuration !== undefined) {
        if (parsedDefaultDuration !== (existingModel.default_duration ?? 0)) {
          payload.default_duration = parsedDefaultDuration;
        }
      } else if (
        !modelForm.default_duration.trim() &&
        (existingModel.default_duration ?? 0) > 0
      ) {
        payload.default_duration = 0;
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
      // New fields comparison
      const trimmedGenerationMode = modelForm.generation_mode.trim();
      if (trimmedGenerationMode !== (existingModel.generation_mode ?? "")) {
        payload.generation_mode = trimmedGenerationMode;
      }
      const trimmedEndpointPath = modelForm.endpoint_path.trim();
      if (trimmedEndpointPath !== (existingModel.endpoint_path ?? "")) {
        payload.endpoint_path = trimmedEndpointPath;
      }
      if (modelForm.supports_streaming !== (existingModel.supports_streaming ?? false)) {
        payload.supports_streaming = modelForm.supports_streaming;
      }
      if (modelForm.supports_cancel !== (existingModel.supports_cancel ?? false)) {
        payload.supports_cancel = modelForm.supports_cancel;
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
        setMessage("模型信息已更新");
        setModelFormOpen(false);
        setModelForm(defaultModelForm());
        setEditingModelId(null);
        setModelFormProviderId(null);
        await loadProviders();
      } catch (err) {
        setModelFormError(
          err instanceof Error ? err.message : "更新模型失败"
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
      modelFormProviderId,
      providers,
    ]
  );

  const handleToggleModelActive = useCallback(
    async (providerId: string, model: ProviderModelAdmin) => {
      const provider =
        providers.find((item) => item.id === providerId) ?? null;
      if (!provider) {
        setError("未找到厂商信息，请重试");
        return;
      }
      clearMessages();
      setBusy(true);
      try {
        await updateProviderModel(provider.id, model.model_id, {
          is_active: !model.is_active,
        });
        setMessage("模型状态已更新");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "更新模型状态失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, providers]
  );

  const handleDeleteModel = useCallback(
    async (providerId: string, model: ProviderModelAdmin) => {
      const provider =
        providers.find((item) => item.id === providerId) ?? null;
      if (!provider) {
        setError("未找到厂商信息，请重试");
        return;
      }
      if (
        !window.confirm(
          `确认删除模型 ${model.model_id} 吗？此操作不可恢复。`
        )
      ) {
        return;
      }
      clearMessages();
      setBusy(true);
      try {
        await deleteProviderModel(provider.id, model.model_id);
        setMessage("模型已删除");
        await loadProviders();
      } catch (err) {
        setError(err instanceof Error ? err.message : "删除模型失败");
      } finally {
        setBusy(false);
      }
    },
    [clearMessages, loadProviders, providers]
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
    <Box sx={{ height: "calc(100vh - 128px)", display: "flex", flexDirection: "column" }}>
      <Box sx={{ mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.5 }}>
          厂商管理
        </Typography>
        <Typography variant="body2" color="text.secondary">
          创建与维护 AI 服务商及其模型配置。
        </Typography>
      </Box>

      {(error || message) && (
        <Stack spacing={2} sx={{ mb: 2 }}>
          {error && <Alert severity="error">{error}</Alert>}
          {message && <Alert severity="success">{message}</Alert>}
        </Stack>
      )}

      {loading && providers.length === 0 ? (
        <Box
          sx={{
            flex: 1,
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <CircularProgress />
        </Box>
      ) : (
        <Box
          sx={{
            flex: 1,
            display: "flex",
            flexDirection: { xs: "column", md: "row" },
            gap: 0,
            overflow: "hidden",
            boxShadow: "0 4px 20px rgba(0,0,0,0.05)",
            borderRadius: 3,
            bgcolor: "background.paper",
            border: "1px solid",
            borderColor: "divider",
          }}
        >
          <ProviderSidebar
            providers={providers}
            selectedProviderId={selectedProviderId}
            onSelect={setSelectedProviderId}
            onAddProvider={openCreateProviderDialog}
          />
          <ProviderDetailPanel
            provider={selectedProvider}
            onEditProvider={openEditProviderDialog}
            onDeleteProvider={handleDeleteProvider}
            onToggleProviderActive={handleToggleProviderActive}
            onAddModel={openCreateModelForm}
            onEditModel={openEditModelForm}
            onDeleteModel={handleDeleteModel}
            onToggleModelActive={handleToggleModelActive}
            onCloneModel={handleCloneModel}
          />
        </Box>
      )}

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
        open={modelFormOpen}
        onClose={closeModelForm}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>
          {modelFormMode === "create"
            ? "新增模型"
            : `编辑模型：${modelForm.model_id}`}
          {modelFormProvider ? `（${modelFormProvider.name}）` : ""}
        </DialogTitle>
        <Box component="form" onSubmit={handleSubmitModelForm}>
          <DialogContent>
            <Stack spacing={2}>
              {modelFormError && <Alert severity="error">{modelFormError}</Alert>}
              {modelFormProvider && (
                <Typography variant="caption" color="text.secondary">
                  归属厂商：{modelFormProvider.name}（{modelFormProvider.id}）
                </Typography>
              )}

              <Grid container spacing={2}>
                {/* Left Column: Basic Info */}
                <Grid size={{ xs: 12, md: 6 }}>
                  <Stack spacing={2}>
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
                        helperText="后端识别标识，如 gpt-4"
                        fullWidth
                      />
                    ) : (
                      <TextField
                        label="模型 ID"
                        value={modelForm.model_id}
                        disabled
                        fullWidth
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
                      fullWidth
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
                      fullWidth
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
                      minRows={3}
                      fullWidth
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
                  </Stack>
                </Grid>

                {/* Right Column: Configuration */}
                <Grid size={{ xs: 12, md: 6 }}>
                  <Stack spacing={2}>
                    <FormControl fullWidth size="small">
                      <InputLabel id="input-modalities-label">
                        输入模态
                      </InputLabel>
                      <Select
                        labelId="input-modalities-label"
                        label="输入模态"
                        multiple
                        value={modelForm.input_modalities}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            input_modalities: normalizeMultiSelect(
                              event.target.value as string | string[]
                            ),
                          }))
                        }
                        renderValue={(selected) =>
                          Array.isArray(selected) && selected.length > 0
                            ? selected.join(", ")
                            : "未选择"
                        }
                      >
                        {modalityOptions.map((option) => (
                          <MenuItem key={option} value={option}>
                            {option}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <FormControl fullWidth size="small">
                      <InputLabel id="output-modalities-label">
                        输出模态
                      </InputLabel>
                      <Select
                        labelId="output-modalities-label"
                        label="输出模态"
                        multiple
                        value={modelForm.output_modalities}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            output_modalities: normalizeMultiSelect(
                              event.target.value as string | string[]
                            ),
                          }))
                        }
                        renderValue={(selected) =>
                          Array.isArray(selected) && selected.length > 0
                            ? selected.join(", ")
                            : "未选择"
                        }
                      >
                        {modalityOptions.map((option) => (
                          <MenuItem key={option} value={option}>
                            {option}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    <TextField
                      label="最大参考图数量"
                      value={modelForm.max_images}
                      onChange={(event) =>
                        setModelForm((prev) => ({
                          ...prev,
                          max_images: event.target.value,
                        }))
                      }
                      placeholder="0"
                      fullWidth
                      size="small"
                    />
                    <Stack direction="row" spacing={1}>
                      <TextField
                        label="默认尺寸"
                        value={modelForm.default_size}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            default_size: event.target.value,
                          }))
                        }
                        fullWidth
                        size="small"
                      />
                      <TextField
                        label="支持尺寸"
                        value={modelForm.supported_sizes}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            supported_sizes: event.target.value,
                          }))
                        }
                        placeholder="1024x1024, ..."
                        fullWidth
                        size="small"
                      />
                    </Stack>
                    <Stack direction="row" spacing={1}>
                      <TextField
                        label="默认时长(s)"
                        value={modelForm.default_duration}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            default_duration: event.target.value,
                          }))
                        }
                        fullWidth
                        size="small"
                      />
                      <TextField
                        label="支持时长"
                        value={modelForm.supported_durations}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            supported_durations: event.target.value,
                          }))
                        }
                        placeholder="5, 10"
                        fullWidth
                        size="small"
                      />
                    </Stack>
                    <TextField
                      label="附加配置 (JSON)"
                      value={modelForm.settings}
                      onChange={(event) =>
                        setModelForm((prev) => ({
                          ...prev,
                          settings: event.target.value,
                        }))
                      }
                      placeholder='{"endpoint": "..."}'
                      multiline
                      minRows={3}
                      fullWidth
                      size="small"
                    />
                    <Stack direction="row" spacing={1}>
                      <TextField
                        select
                        label="生成模式"
                        value={modelForm.generation_mode}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            generation_mode: event.target.value,
                          }))
                        }
                        fullWidth
                        size="small"
                      >
                        <MenuItem value="">未选择</MenuItem>
                        {generationModeOptions.map((option) => (
                          <MenuItem key={option} value={option}>
                            {option}
                          </MenuItem>
                        ))}
                      </TextField>
                      <TextField
                        label="端点路径"
                        value={modelForm.endpoint_path}
                        onChange={(event) =>
                          setModelForm((prev) => ({
                            ...prev,
                            endpoint_path: event.target.value,
                          }))
                        }
                        placeholder="/v1/images/generations"
                        fullWidth
                        size="small"
                      />
                    </Stack>
                    <Stack direction="row" spacing={2}>
                      <FormControlLabel
                        control={
                          <Switch
                            checked={modelForm.supports_streaming}
                            onChange={(event) =>
                              setModelForm((prev) => ({
                                ...prev,
                                supports_streaming: event.target.checked,
                              }))
                            }
                            size="small"
                          />
                        }
                        label="支持流式输出"
                      />
                      <FormControlLabel
                        control={
                          <Switch
                            checked={modelForm.supports_cancel}
                            onChange={(event) =>
                              setModelForm((prev) => ({
                                ...prev,
                                supports_cancel: event.target.checked,
                              }))
                            }
                            size="small"
                          />
                        }
                        label="支持取消任务"
                      />
                    </Stack>
                  </Stack>
                </Grid>
              </Grid>
            </Stack>
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
