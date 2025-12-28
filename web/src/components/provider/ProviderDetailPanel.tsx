import React from "react";
import {
    Box,
    Button,
    Chip,
    Grid,
    IconButton,
    Paper,
    Stack,
    Tooltip,
    Typography,
    Card,
    CardContent,
} from "@mui/material";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteForeverRoundedIcon from "@mui/icons-material/DeleteForeverRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import ToggleOnRoundedIcon from "@mui/icons-material/ToggleOnRounded";
import ToggleoffRoundedIcon from "@mui/icons-material/ToggleOffRounded";
import ContentCopyRoundedIcon from "@mui/icons-material/ContentCopyRounded";
import Inventory2OutlinedIcon from "@mui/icons-material/Inventory2Outlined";
import ContentPasteGoRoundedIcon from "@mui/icons-material/ContentPasteGoRounded";

import type { ProviderAdmin, ProviderModelAdmin } from "../../types";

interface ProviderDetailPanelProps {
    provider: ProviderAdmin | null;
    onEditProvider: (provider: ProviderAdmin) => void;
    onDeleteProvider: (provider: ProviderAdmin) => void;
    onToggleProviderActive: (provider: ProviderAdmin) => void;
    onAddModel: (providerId: string) => void;
    onEditModel: (providerId: string, model: ProviderModelAdmin) => void;
    onDeleteModel: (providerId: string, model: ProviderModelAdmin) => void;
    onToggleModelActive: (providerId: string, model: ProviderModelAdmin) => void;
    onCloneModel: (providerId: string, model: ProviderModelAdmin) => void;
}

const ProviderDetailPanel: React.FC<ProviderDetailPanelProps> = ({
    provider,
    onEditProvider,
    onDeleteProvider,
    onToggleProviderActive,
    onAddModel,
    onEditModel,
    onDeleteModel,
    onToggleModelActive,
    onCloneModel,
}) => {
    if (!provider) {
        return (
            <Box
                sx={{
                    flex: 1,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    bgcolor: "background.default",
                    borderRadius: { xs: 2, md: "0 12px 12px 0" },
                    minHeight: 400,
                }}
            >
                <Stack alignItems="center" spacing={2} sx={{ color: "text.secondary" }}>
                    <Inventory2OutlinedIcon sx={{ fontSize: 48, opacity: 0.5 }} />
                    <Typography>请选择一个厂商以管理</Typography>
                </Stack>
            </Box>
        );
    }

    const activeModels = provider.models?.filter((m) => m.is_active).length ?? 0;

    return (
        <Box
            sx={{
                flex: 1,
                display: "flex",
                flexDirection: "column",
                bgcolor: "background.default",
                borderRadius: { xs: 2, md: "0 12px 12px 0" },
                overflow: "hidden",
            }}
        >
            {/* Header Section */}
            <Paper
                elevation={0}
                sx={{
                    p: 2,
                    borderBottom: "1px solid",
                    borderColor: "divider",
                    borderRadius: 0,
                }}
            >
                <Stack
                    direction={{ xs: "column", sm: "row" }}
                    justifyContent="space-between"
                    alignItems={{ xs: "flex-start", sm: "flex-start" }}
                    spacing={2}
                >
                    <Box sx={{ flex: 1 }}>
                        <Stack direction="row" alignItems="center" spacing={1.5} sx={{ mb: 1 }}>
                            <Typography variant="h5" sx={{ fontWeight: 700 }}>
                                {provider.name}
                            </Typography>
                            <Chip
                                label={provider.is_active ? "启用中" : "已停用"}
                                color={provider.is_active ? "success" : "default"}
                                size="small"
                                variant="filled"
                                sx={{ borderRadius: 1.5, fontWeight: 600 }}
                            />
                            <Chip
                                label={provider.driver}
                                size="small"
                                variant="outlined"
                                sx={{ borderRadius: 1.5, fontFamily: "monospace" }}
                            />
                        </Stack>
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ mb: 1.5, maxWidth: "800px" }}
                        >
                            {provider.description || "暂无描述"}
                        </Typography>
                        <Stack direction="row" spacing={3} alignItems="center">
                            <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                                <Typography variant="caption" color="text.secondary">
                                    ID:
                                </Typography>
                                <Typography variant="caption" sx={{ fontFamily: "monospace" }}>
                                    {provider.id}
                                </Typography>
                                <IconButton size="small" sx={{ p: 0.5 }}>
                                    <ContentCopyRoundedIcon sx={{ fontSize: 14 }} />
                                </IconButton>
                            </Box>
                            <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                                <Typography variant="caption" color="text.secondary">
                                    接口:
                                </Typography>
                                <Typography variant="caption" sx={{ fontFamily: "monospace" }}>
                                    {provider.base_url || "默认"}
                                </Typography>
                            </Box>
                        </Stack>
                    </Box>
                    <Stack direction="row" spacing={1}>
                        <Button
                            variant="outlined"
                            size="small"
                            startIcon={<EditRoundedIcon />}
                            onClick={() => onEditProvider(provider)}
                        >
                            编辑
                        </Button>
                        <Button
                            variant={provider.is_active ? "text" : "outlined"}
                            size="small"
                            color={provider.is_active ? "warning" : "success"}
                            onClick={() => onToggleProviderActive(provider)}
                        >
                            {provider.is_active ? "停用" : "启用"}
                        </Button>
                        <Tooltip title="删除整组厂商配置">
                            <IconButton
                                color="error"
                                size="small"
                                onClick={() => onDeleteProvider(provider)}
                                sx={{
                                    border: '1px solid',
                                    borderColor: 'error.light',
                                    '&:hover': { bgcolor: 'error.soft' }
                                }}
                            >
                                <DeleteForeverRoundedIcon fontSize="small" />
                            </IconButton>
                        </Tooltip>
                    </Stack>
                </Stack>
            </Paper>

            {/* Models Section */}
            <Box sx={{ p: 2, flex: 1, overflowY: "auto" }}>
                <Stack
                    direction="row"
                    alignItems="center"
                    justifyContent="space-between"
                    sx={{ mb: 2 }}
                >
                    <Typography variant="h6" sx={{ fontWeight: 700 }}>
                        模型配置
                        <Typography
                            component="span"
                            variant="body2"
                            color="text.secondary"
                            sx={{ ml: 1.5, fontWeight: 400 }}
                        >
                            共 {provider.models?.length ?? 0} 个模型，启用 {activeModels} 个
                        </Typography>
                    </Typography>
                    <Button
                        variant="contained"
                        size="small"
                        startIcon={<AddRoundedIcon />}
                        onClick={() => onAddModel(provider.id)}
                    >
                        新增模型
                    </Button>
                </Stack>

                {!provider.models || provider.models.length === 0 ? (
                    <Paper
                        elevation={0}
                        sx={{
                            p: 4,
                            textAlign: "center",
                            border: "1px dashed",
                            borderColor: "divider",
                            bgcolor: "transparent",
                        }}
                    >
                        <Typography color="text.secondary">暂无模型配置</Typography>
                        <Button
                            variant="text"
                            startIcon={<AddRoundedIcon />}
                            onClick={() => onAddModel(provider.id)}
                            sx={{ mt: 1 }}
                        >
                            立即添加
                        </Button>
                    </Paper>
                ) : (
                    <Grid container spacing={2}>
                        {provider.models.map((model) => (
                            <Grid size={12} key={model.model_id}>
                                <Card
                                    elevation={0}
                                    sx={{
                                        border: "1px solid",
                                        borderColor: "divider",
                                        transition: "all 0.2s",
                                        "&:hover": {
                                            borderColor: "primary.light",
                                            boxShadow: "0 4px 12px rgba(0,0,0,0.05)",
                                        },
                                        opacity: model.is_active ? 1 : 0.7,
                                    }}
                                >
                                    <CardContent sx={{ p: 2, "&:last-child": { pb: 2 } }}>
                                        <Grid container spacing={2} alignItems="center">
                                            <Grid size={{ xs: 12, md: 4 }}>
                                                <Stack spacing={0.5}>
                                                    <Stack direction="row" alignItems="center" spacing={1}>
                                                        <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
                                                            {model.name}
                                                        </Typography>
                                                        <Chip
                                                            label={model.model_id}
                                                            size="small"
                                                            sx={{
                                                                fontSize: "0.7rem",
                                                                height: 20,
                                                                bgcolor: "action.hover",
                                                                fontFamily: "monospace"
                                                            }}
                                                        />
                                                    </Stack>
                                                    {model.description && (
                                                        <Typography variant="body2" color="text.secondary" noWrap>
                                                            {model.description}
                                                        </Typography>
                                                    )}
                                                    {model.price && (
                                                        <Typography variant="caption" color="text.secondary">
                                                            💰 {model.price}
                                                        </Typography>
                                                    )}
                                                </Stack>
                                            </Grid>
                                            <Grid size={{ xs: 12, md: 5 }}>
                                                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ gap: 0.5 }}>
                                                    {model.input_modalities?.map(m => (
                                                        <Chip key={m} label={`In:${m} `} size="small" variant="outlined" sx={{ borderRadius: 1 }} />
                                                    ))}
                                                    {model.output_modalities?.map(m => (
                                                        <Chip key={m} label={`Out:${m} `} size="small" color="primary" variant="outlined" sx={{ borderRadius: 1 }} />
                                                    ))}
                                                </Stack>
                                            </Grid>
                                            <Grid size={{ xs: 12, md: 3 }}>
                                                <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                                                    <Tooltip title="克隆模型">
                                                        <IconButton
                                                            size="small"
                                                            onClick={() => onCloneModel(provider.id, model)}
                                                        >
                                                            <ContentPasteGoRoundedIcon fontSize="small" />
                                                        </IconButton>
                                                    </Tooltip>
                                                    <Tooltip title={model.is_active ? "停用" : "启用"}>
                                                        <IconButton
                                                            size="small"
                                                            color={model.is_active ? "success" : "default"}
                                                            onClick={() => onToggleModelActive(provider.id, model)}
                                                        >
                                                            {model.is_active ? <ToggleOnRoundedIcon /> : <ToggleoffRoundedIcon />}
                                                        </IconButton>
                                                    </Tooltip>
                                                    <Tooltip title="编辑">
                                                        <IconButton size="small" onClick={() => onEditModel(provider.id, model)}>
                                                            <EditRoundedIcon fontSize="small" />
                                                        </IconButton>
                                                    </Tooltip>
                                                    <Tooltip title="删除">
                                                        <IconButton
                                                            size="small"
                                                            color="error"
                                                            onClick={() => onDeleteModel(provider.id, model)}
                                                        >
                                                            <DeleteForeverRoundedIcon fontSize="small" />
                                                        </IconButton>
                                                    </Tooltip>
                                                </Stack>
                                            </Grid>
                                        </Grid>
                                    </CardContent>
                                </Card>
                            </Grid>
                        ))}
                    </Grid>
                )}
            </Box>
        </Box>
    );
};

export default ProviderDetailPanel;
